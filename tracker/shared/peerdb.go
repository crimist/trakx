package shared

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

type PeerDatabase struct {
	mu sync.RWMutex
	db map[Hash]map[PeerID]Peer

	conf   *Config
	logger *zap.Logger
}

func (db *PeerDatabase) check() (ok bool) {
	if db.db != nil {
		ok = true
	}
	return
}

// Trim removes all peers that haven't checked in since timeout
func (db *PeerDatabase) Trim() {
	db.logger.Info("Trimming database")

	start := time.Now()
	var peers, hashes int
	now := time.Now().Unix()

	db.mu.Lock()
	for hash, peermap := range db.db {
		for id, peer := range peermap {
			if now-peer.LastSeen > db.conf.Database.Peer.Timeout {
				db.deletePeer(&peer, &hash, &id)
				db.deleteIP(peer.IP)
				peers++
			}
		}
		if len(peermap) == 0 {
			delete(db.db, hash)
			hashes++
		}
	}
	db.mu.Unlock()

	db.logger.Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *PeerDatabase) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&PeerDB.db)
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
				PeerDB = PeerDatabase{db: make(map[Hash]map[PeerID]Peer)}
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
				PeerDB = PeerDatabase{db: make(map[Hash]map[PeerID]Peer)}
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
				PeerDB = PeerDatabase{db: make(map[Hash]map[PeerID]Peer)}
				return
			} else {
				loaded = "temp"
			}
		} else {
			loaded = "full"
		}
	}

	db.logger.Info("Loaded database", zap.String("type", loaded), zap.Int("hashes", PeerDB.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
}

func (db *PeerDatabase) write(temp bool) {
	db.logger.Info("Writing database")

	start := time.Now()
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)
	filename := db.conf.Database.Peer.Filename
	if temp {
		filename += ".tmp"
	} else {
		db.Trim()
	}

	db.mu.RLock()
	err := encoder.Encode(&PeerDB.db)
	db.mu.RUnlock()
	if err != nil {
		db.logger.Error("peerdb gob encoder", zap.Error(err))
		return
	}

	if err := ioutil.WriteFile(filename, buff.Bytes(), 0644); err != nil {
		db.logger.Error("peerdb writefile", zap.Error(err))
		return
	}

	db.logger.Info("Wrote database", zap.String("filename", filename), zap.Int("hashes", PeerDB.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
}

// WriteTmp writes the database to tmp file
func (db *PeerDatabase) WriteTmp() {
	db.write(true)
}

// WriteFull writes the database to file
func (db *PeerDatabase) WriteFull() {
	db.write(false)
}
