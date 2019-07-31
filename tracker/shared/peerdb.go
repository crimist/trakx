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
	mu sync.RWMutex
	db map[Hash]map[PeerID]Peer
}

func (db *PeerDatabase) check() (ok bool) {
	if db.db != nil {
		ok = true
	}
	return
}

// Hashes gets the number of hashes
func (db *PeerDatabase) Hashes() int {
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
			ExpvarIPs.Lock()
			delete(ExpvarIPs.M, dbPeer.IP)
			ExpvarIPs.M[p.IP]++
			ExpvarIPs.Unlock()
		}
	} else { // New
		ExpvarIPs.Lock()
		ExpvarIPs.M[p.IP]++
		ExpvarIPs.Unlock()
		if p.Complete {
			atomic.AddInt64(&ExpvarSeeds, 1)
		} else {
			atomic.AddInt64(&ExpvarLeeches, 1)
		}
	}

	db.db[h][id] = *p
	db.mu.Unlock()
}

// Drop deletes peer
func (db *PeerDatabase) Drop(p *Peer, h Hash, id PeerID) {
	if !Config.Trakx.Prod {
		Logger.Info("Drop",
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

	ExpvarIPs.Lock()
	ExpvarIPs.M[p.IP]--
	if ExpvarIPs.M[p.IP] < 1 {
		delete(ExpvarIPs.M, p.IP)
	}
	ExpvarIPs.Unlock()

	delete(db.db[h], id)
	db.mu.Unlock()
}

// Delete deletes a hash
func (db *PeerDatabase) delete(h Hash) {
	db.mu.Lock()
	delete(db.db, h)
	db.mu.Unlock()
}

// Trim removes all peers that haven't checked in since timeout
func (db *PeerDatabase) Trim() {
	var peers, hashes int
	now := time.Now().Unix()

	db.mu.RLock()
	for hash, peermap := range db.db {
		for id, peer := range peermap {
			if now-peer.LastSeen > Config.Database.Peer.Timeout {
				db.Drop(&peer, hash, id)
				peers++
			}
		}
		if len(peermap) == 0 {
			db.delete(hash)
			hashes++
		}
	}
	db.mu.RUnlock()

	Logger.Info("Trimmed PeerDatabase", zap.Int("peers", peers), zap.Int("hashes", hashes))
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

	if err := encoder.Encode(&PeerDB.db); err != nil {
		Logger.Error("peerdb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(filename, buff.Bytes(), 0644); err != nil {
		Logger.Error("peerdb writefile", zap.Error(err))
	}

	Logger.Info("Wrote PeerDatabase", zap.String("filename", filename), zap.Int("hashes", PeerDB.Hashes()))
}

// WriteTmp writes the database to tmp file
func (db *PeerDatabase) WriteTmp() {
	db.write(true)
}

// WriteFull writes	 the database to file
func (db *PeerDatabase) WriteFull() {
	db.write(false)
}
