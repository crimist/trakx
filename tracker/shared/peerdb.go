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
}

func (db *PeerDatabase) check() (ok bool) {
	if db.db != nil {
		ok = true
	}
	return
}

// Trim removes all peers that haven't checked in since timeout
func (db *PeerDatabase) Trim() {
	start := time.Now()
	Logger.Info("Trimming database")
	var peers, hashes int
	now := time.Now().Unix()

	db.mu.Lock()
	for hash, peermap := range db.db {
		for id, peer := range peermap {
			if now-peer.LastSeen > Config.Database.Peer.Timeout {
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

	Logger.Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Now().Sub(start)))
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
	loadtemp := false

	infoFull, err := os.Stat(Config.Database.Peer.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("No full peerdb")
			loadtemp = true
		} else {
			Logger.Error("os.Stat", zap.Error(err))
		}
	}
	infoTemp, err := os.Stat(Config.Database.Peer.Filename + ".tmp")
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("No temp peerdb")
			if loadtemp {
				Logger.Info("No peerdb found")
				PeerDB = PeerDatabase{db: make(map[Hash]map[PeerID]Peer)}
				return
			}
		} else {
			Logger.Error("os.Stat", zap.Error(err))
		}
	}

	if infoFull != nil && infoTemp != nil {
		if infoTemp.ModTime().UnixNano() > infoFull.ModTime().UnixNano() {
			loadtemp = true
		}
	}

	loaded := ""
	if loadtemp == true {
		if err := db.load(Config.Database.Peer.Filename + ".tmp"); err != nil {
			Logger.Info("Loading temp peerdb failed", zap.Error(err))

			if err := db.load(Config.Database.Peer.Filename); err != nil {
				Logger.Info("Loading full peerdb failed", zap.Error(err))
				PeerDB = PeerDatabase{db: make(map[Hash]map[PeerID]Peer)}
				return
			} else {
				loaded = "full"
			}
		} else {
			loaded = "temp"
		}
	} else {
		if err := db.load(Config.Database.Peer.Filename); err != nil {
			Logger.Info("Loading full peerdb failed", zap.Error(err))

			if err := db.load(Config.Database.Peer.Filename + ".tmp"); err != nil {
				Logger.Info("Loading temp peerdb failed", zap.Error(err))
				PeerDB = PeerDatabase{db: make(map[Hash]map[PeerID]Peer)}
				return
			} else {
				loaded = "temp"
			}
		} else {
			loaded = "full"
		}
	}

	Logger.Info("Loaded peerdb", zap.String("type", loaded), zap.Int("hashes", PeerDB.Hashes()))
}

func (db *PeerDatabase) write(temp bool) {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)
	filename := Config.Database.Peer.Filename
	if temp {
		filename += ".tmp"
	} else {
		db.Trim()
	}

	db.mu.RLock()
	err := encoder.Encode(&PeerDB.db)
	db.mu.RUnlock()
	if err != nil {
		Logger.Error("peerdb gob encoder", zap.Error(err))
		return
	}

	if err := ioutil.WriteFile(filename, buff.Bytes(), 0644); err != nil {
		Logger.Error("peerdb writefile", zap.Error(err))
		return
	}

	Logger.Info("Wrote PeerDatabase", zap.String("filename", filename), zap.Int("hashes", PeerDB.Hashes()))
}

// WriteTmp writes the database to tmp file
func (db *PeerDatabase) WriteTmp() {
	db.write(true)
}

// WriteFull writes the database to file
func (db *PeerDatabase) WriteFull() {
	db.write(false)
}
