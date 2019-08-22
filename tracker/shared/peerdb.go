package shared

import (
	"archive/zip" // TODO use gzip
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"io/ioutil"
	"os"
	"sync"
	"time"

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

	peerdb.Load()

	go RunOn(time.Duration(conf.Database.Peer.Write)*time.Second, peerdb.WriteTmp)
	go RunOn(time.Duration(conf.Database.Peer.Trim)*time.Second, peerdb.Trim)

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

// Load loads a database into memory
func (db *PeerDatabase) Load() {
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
		if err := db.load(db.conf.Database.Peer.Filename + ".tmp"); err != nil {
			db.logger.Info("Loading temp peerdb failed", zap.Error(err))

			if err := db.load(db.conf.Database.Peer.Filename); err != nil {
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
		if err := db.load(db.conf.Database.Peer.Filename); err != nil {
			db.logger.Info("Loading full peerdb failed", zap.Error(err))

			if err := db.load(db.conf.Database.Peer.Filename + ".tmp"); err != nil {
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

func (db *PeerDatabase) load(filename string) error {
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

// like trim() this uses costly locking but it's worth it to prevent blocking
func (db *PeerDatabase) write(temp bool) bool {
	var buff bytes.Buffer
	archive := zip.NewWriter(&buff)
	filename := db.conf.Database.Peer.Filename
	if temp {
		filename += ".tmp"
	} else {
		db.Trim()
	}

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
		return false
	}

	db.logger.Info("Writing zip to file", zap.Float32("mb", float32(buff.Len())/1024.0/1024.0))
	if err := ioutil.WriteFile(filename, buff.Bytes(), 0644); err != nil {
		db.logger.Error("Database writefile failed", zap.Error(err))
		return false
	}
	return true
}

// WriteTmp writes the database to tmp file
func (db *PeerDatabase) WriteTmp() {
	db.logger.Info("Writing temp database")
	start := time.Now()
	if !db.write(true) {
		db.logger.Info("Failed to write temp database", zap.Duration("duration", time.Now().Sub(start)))
		return
	}
	db.logger.Info("Wrote temp database", zap.Int("hashes", db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
}

// WriteFull writes the database to file
func (db *PeerDatabase) WriteFull() {
	db.logger.Info("Writing full database")
	start := time.Now()
	if !db.write(false) {
		db.logger.Info("Failed to write full database", zap.Duration("duration", time.Now().Sub(start)))
		return
	}
	db.logger.Info("Wrote fuill database", zap.Int("hashes", db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
}
