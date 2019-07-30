package shared

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type PeerDatabase struct {
	mu sync.Mutex
	db map[Hash]map[PeerID]Peer
}

func (db *PeerDatabase) Len() int {
	return len(db.db)
}

// Save creates or updates peer
func (db *PeerDatabase) Save(p *Peer, h Hash, id PeerID) {
	if !Config.Trakx.Prod {
		Logger.Info("Save",
			zap.Any("hash", h),
			zap.Any("peerid", id),
			zap.Any("Peer", p),
		)
	}

	db.mu.Lock()

	// Create map if it doesn't exist
	if _, ok := db.db[h]; !ok {
		db.db[h] = make(map[PeerID]Peer)
		if !Config.Trakx.Prod {
			Logger.Info("Created hash map", zap.Any("hash", h[:]))
		}
	}

	dbPeer, ok := db.db[h][id]
	if ok { // Already in db
		if dbPeer.Complete == false && p.Complete == true { // They completed
			atomic.AddInt64(&ExpvarLeeches, -1)
			atomic.AddInt64(&ExpvarSeeds, 1)
		}
		if dbPeer.Complete == true && p.Complete == false { // They uncompleted?
			atomic.AddInt64(&ExpvarSeeds, -1)
			atomic.AddInt64(&ExpvarLeeches, 1)
		}
		if dbPeer.IP != p.IP { // IP changed
			delete(ExpvarIPs, dbPeer.IP)
			ExpvarIPs[p.IP]++
		}
	} else { // New
		ExpvarIPs[p.IP]++
		if p.Complete {
			atomic.AddInt64(&ExpvarSeeds, 1)
		} else {
			atomic.AddInt64(&ExpvarLeeches, 1)
		}
	}

	db.db[h][id] = *p
	db.mu.Unlock()
}

// Delete deletes peer
func (db *PeerDatabase) Delete(p *Peer, h Hash, id PeerID) {
	if !Config.Trakx.Prod {
		Logger.Info("Delete",
			zap.Any("hash", h),
			zap.Any("peerid", id),
			zap.Any("Peer", p),
		)
	}

	db.mu.Lock()

	if peer, ok := db.db[h][id]; ok {
		if peer.Complete {
			atomic.AddInt64(&ExpvarSeeds, -1)
		} else {
			atomic.AddInt64(&ExpvarLeeches, -1)
		}
	}

	ExpvarIPs[p.IP]--
	if ExpvarIPs[p.IP] < 1 {
		delete(ExpvarIPs, p.IP)
	}

	delete(db.db[h], id)
	db.mu.Unlock()
}

// Trim removes all peers that haven't checked in since timeout
func (db *PeerDatabase) Trim() {
	var peers, hashes int
	now := time.Now().Unix()

	for hash, peermap := range db.db {
		for id, peer := range peermap {
			if now-peer.LastSeen > Config.Database.Peer.Timeout {
				db.Delete(&peer, hash, id)
				peers++
			}
		}
		if len(peermap) == 0 {
			db.mu.Lock()
			delete(db.db, hash)
			db.mu.Unlock()
			hashes++
		}
	}

	Logger.Info("Trimmed PeerDatabase", zap.Int("peers", peers), zap.Int("hashes", hashes))
}

func (db *PeerDatabase) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&PeerDB)
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

	Logger.Info("Loaded peerdb", zap.String("type", loaded), zap.Int("hashes", PeerDB.Len()))
}

// WriteTmp dumps the database to the tmp file
func (db *PeerDatabase) WriteTmp() {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(&PeerDB); err != nil {
		Logger.Error("peerdb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(Config.Database.Peer.Filename+".tmp", buff.Bytes(), 0644); err != nil {
		Logger.Error("peerdb writefile", zap.Error(err))
	}

	Logger.Info("Wrote temp peerdb", zap.Int("hashes", PeerDB.Len()))
}

// WriteFull dumps the database to the db file
func (db *PeerDatabase) WriteFull() {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	db.Trim() // trim to remove nil refs

	if err := encoder.Encode(&PeerDB); err != nil {
		Logger.Error("peerdb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(Config.Database.Peer.Filename, buff.Bytes(), 0644); err != nil {
		Logger.Error("peerdb writefile", zap.Error(err))
	}

	Logger.Info("Wrote full peerdb", zap.Int("hashes", PeerDB.Len()))
}
