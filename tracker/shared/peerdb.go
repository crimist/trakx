package shared

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	peerdbHashCap = 1000000
)

type PeerMap struct {
	sync.RWMutex
	peers map[PeerID]*Peer
}

type PeerDatabase struct {
	mu      sync.RWMutex
	hashmap map[Hash]*PeerMap

	conf   *Config
	logger *zap.Logger
}

func NewPeerDatabase(conf *Config, logger *zap.Logger) *PeerDatabase {
	peerdb := PeerDatabase{
		conf:   conf,
		logger: logger,
	}

	peerdb.LoadFromFile()

	if conf.Database.Peer.Write > 0 {
		go RunOn(time.Duration(conf.Database.Peer.Write)*time.Second, peerdb.WriteTmp)
	}
	if conf.Database.Peer.Trim > 0 {
		go RunOn(time.Duration(conf.Database.Peer.Trim)*time.Second, peerdb.Trim)
	}

	return &peerdb
}

func (db *PeerDatabase) check() (ok bool) {
	if db.hashmap != nil {
		ok = true
	}
	return
}

func (db *PeerDatabase) make() {
	db.hashmap = make(map[Hash]*PeerMap, peerdbHashCap)
}

// Trim removes all peers that haven't checked in since timeout
func (db *PeerDatabase) Trim() {
	start := time.Now()
	db.logger.Info("Trimming database")
	peers, hashes := db.trim()
	if peers < 1 && hashes < 1 {
		db.logger.Info("Can't trim database: database empty")
	} else {
		db.logger.Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Now().Sub(start)))
	}
}

// the spam lock and unlock is expensive but it stops the program from blocks for seconds
// this is especially important on a single core slow system
func (db *PeerDatabase) trim() (peers, hashes int) {
	now := time.Now().Unix()

	db.mu.Lock()
	for hash, peermap := range db.hashmap {
		db.mu.Unlock()

		peermap.Lock()
		for id, peer := range peermap.peers {
			if now-peer.LastSeen > db.conf.Database.Peer.Timeout {
				db.delete(peer, &hash, &id)
				peers++
			}
		}
		peermap.Unlock()

		db.mu.Lock()
		if len(peermap.peers) == 0 {
			delete(db.hashmap, hash)
			hashes++
		}
	}
	db.mu.Unlock()

	return
}

func (db *PeerDatabase) LoadFromFile() {
	db.logger.Info("Loading database")
	start := time.Now()
	loadtemp := false

	infoFull, err := os.Stat(db.conf.Database.Peer.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			db.logger.Info("No full peerdb")
			loadtemp = true
		} else {
			db.logger.Error("os.Stat", zap.Error(err))
		}
	}
	infoTemp, err := os.Stat(db.conf.Database.Peer.Filename + ".tmp")
	if err != nil {
		if os.IsNotExist(err) {
			db.logger.Info("No temp peerdb")
			if loadtemp {
				db.logger.Info("No peerdb found")
				db.make()
				return
			}
		} else {
			db.logger.Error("os.Stat", zap.Error(err))
		}
	}

	if infoFull != nil && infoTemp != nil {
		if infoTemp.ModTime().UnixNano() > infoFull.ModTime().UnixNano() {
			loadtemp = true
		}
	}

	loaded := ""
	if loadtemp == true {
		if err := db.loadFile(db.conf.Database.Peer.Filename + ".tmp"); err != nil {
			db.logger.Info("Loading temp peerdb failed", zap.Error(err))

			if err := db.loadFile(db.conf.Database.Peer.Filename); err != nil {
				db.logger.Info("Loading full peerdb failed", zap.Error(err))
				db.make()
				return
			} else {
				loaded = "full"
			}
		} else {
			loaded = "temp"
		}
	} else {
		if err := db.loadFile(db.conf.Database.Peer.Filename); err != nil {
			db.logger.Info("Loading full peerdb failed", zap.Error(err))

			if err := db.loadFile(db.conf.Database.Peer.Filename + ".tmp"); err != nil {
				db.logger.Info("Loading temp peerdb failed", zap.Error(err))
				db.make()
				return
			} else {
				loaded = "temp"
			}
		} else {
			loaded = "full"
		}
	}

	db.logger.Info("Loaded database", zap.String("type", loaded), zap.Int("hashes", db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *PeerDatabase) loadFile(filename string) error {
	var hash Hash
	db.make()

	archive, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		hashbytes, err := hex.DecodeString(file.Name)
		if err != nil {
			return err
		}
		copy(hash[:], hashbytes)
		peermap := db.makePeermap(&hash)

		reader, err := file.Open()
		if err != nil {
			return err
		}
		err = gob.NewDecoder(reader).Decode(&peermap.peers)
		if err != nil {
			return err
		}
		reader.Close()
	}

	return nil
}

func (db *PeerDatabase) encode() []byte {
	var buff bytes.Buffer
	archive := zip.NewWriter(&buff)

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		submap.RLock()
		writer, err := archive.Create(hex.EncodeToString(hash[:]))
		if err != nil {
			db.logger.Error("Failed to create in archive", zap.Error(err), zap.Any("hash", hash[:]))
			submap.RUnlock()
			continue
		}
		if err := gob.NewEncoder(writer).Encode(submap.peers); err != nil {
			db.logger.Warn("Failed to encode a peermap", zap.Error(err), zap.Any("hash", hash[:]))
		}
		submap.RUnlock()

		db.mu.RLock()
	}
	db.mu.RUnlock()

	if err := archive.Close(); err != nil {
		db.logger.Error("Failed to close archive", zap.Error(err))
		return nil
	}

	return buff.Bytes()
}

// like trim() this uses costly locking but it's worth it to prevent blocking
func (db *PeerDatabase) writeToFile(temp bool) int {
	filename := db.conf.Database.Peer.Filename
	if temp {
		filename += ".tmp"
	} else {
		db.Trim()
	}

	encoded := db.encode()
	if encoded == nil {
		return -1
	}

	db.logger.Info("Writing zip to file", zap.Float32("mb", float32(len(encoded)/1024.0/1024.0)))
	if err := ioutil.WriteFile(filename, encoded, 0644); err != nil {
		db.logger.Error("Database writefile failed", zap.Error(err))
		return -1
	}
	return len(encoded)
}

// WriteTmp writes the database to tmp file
func (db *PeerDatabase) WriteTmp() {
	db.logger.Info("Writing temp database")
	start := time.Now()
	if db.writeToFile(true) == -1 {
		db.logger.Info("Failed to write temp database", zap.Duration("duration", time.Now().Sub(start)))
		return
	}
	db.logger.Info("Wrote temp database", zap.Int("hashes", db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
}

// WriteFull writes the database to file
func (db *PeerDatabase) WriteFull() {
	db.logger.Info("Writing full database")
	start := time.Now()
	if db.writeToFile(false) == -1 {
		db.logger.Info("Failed to write full database", zap.Duration("duration", time.Now().Sub(start)))
		return
	}
	db.logger.Info("Wrote full database", zap.Int("hashes", db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *PeerDatabase) LoadFromDB() {
	pgdb, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()
	err = pgdb.Ping()
	if err != nil {
		db.logger.Error("postgres ping() failed", zap.Error(err))
		return
	}

	var data []byte
	err = pgdb.QueryRow("SELECT bytes FROM peerdb ORDER BY ts DESC LIMIT 1").Scan(&data)
	if err != nil {
		db.logger.Error("postgres select failed", zap.Error(err))
		return
	}
	fmt.Printf("%s\n", data)
}

func (db *PeerDatabase) WriteToDB() int {
	pgdb, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()
	err = pgdb.Ping()
	if err != nil {
		db.logger.Error("postgres ping() failed", zap.Error(err))
		return -1
	}

	_, err = pgdb.Exec("CREATE TABLE IF NOT EXISTS peerdb (ts TIMESTAMP DEFAULT now(), bytes TEXT)")
	if err != nil {
		db.logger.Error("postgres table create failed", zap.Error(err))
		return -1
	}

	data := db.encode()
	if data == nil {
		db.logger.Error("Failed to encode db")
		return -1
	}

	_, err = pgdb.Query("INSERT INTO peerdb(bytes) VALUES($1)", data)
	if err != nil {
		db.logger.Error("postgres insert failed", zap.Error(err))
		return -1
	}

	rm, err := trimBackups(pgdb)
	if err != nil {
		db.logger.Error("failed to trim backups", zap.Error(err))
		return -1
	}
	db.logger.Info("Deleted expired postgres records", zap.Int64("deleted", rm))

	return len(data)
}

func trimBackups(db *sql.DB) (int64, error) {
	// delete records older than 7 days
	result, err := db.Exec("DELETE FROM peerdb WHERE ts < NOW() - INTERVAL '7 days'")
	if err != nil {
		return -1, err
	}
	rm, err := result.RowsAffected()
	if err != nil {
		return -1, err
	}

	return rm, nil
}
