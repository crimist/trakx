package database

import (
	"sync"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

const initCap = 1000000

type subPeerMap struct {
	sync.RWMutex
	peers map[shared.PeerID]*shared.Peer
}

type Memory struct {
	mu      sync.RWMutex
	hashmap map[shared.Hash]*subPeerMap

	backup MemoryBackup
	conf   *shared.Config
	logger *zap.Logger
}

func (db *Memory) Backup() Backup {
	return db.backup
}

func NewMemory(conf *shared.Config, logger *zap.Logger) *Memory {
	db := Memory{
		conf:   conf,
		logger: logger,
	}

	db.backup.Load()

	if conf.Database.Peer.Write > 0 {
		// go shared.RunOn(time.Duration(conf.Database.Peer.Write)*time.Second, db.backup.Save())
	}
	if conf.Database.Peer.Trim > 0 {
		go shared.RunOn(time.Duration(conf.Database.Peer.Trim)*time.Second, db.Trim)
	}

	return &db
}

func (db *Memory) Check() (ok bool) {
	if db.hashmap != nil {
		ok = true
	}
	return
}

func (db *Memory) Trim() {
	start := time.Now()
	db.logger.Info("Trimming database")
	peers, hashes := db.trim()
	if peers < 1 && hashes < 1 {
		db.logger.Info("Can't trim database: database empty")
	} else {
		db.logger.Info("Trimmed database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("duration", time.Now().Sub(start)))
	}
}

func (db *Memory) trim() (peers, hashes int) {
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

func (db *Memory) makePeermap(h *shared.Hash) (peermap *subPeerMap) {
	// build struct and assign
	peermap = new(subPeerMap)
	peermap.peers = make(map[shared.PeerID]*shared.Peer, 1)
	db.hashmap[*h] = peermap
	return
}

// Save writes a peer
func (db *Memory) Save(p *shared.Peer, h *shared.Hash, id *shared.PeerID) {
	var dbPeer *shared.Peer

	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		db.mu.Lock()
		peermap = db.makePeermap(h)
		db.mu.Unlock()
	}

	peermap.Lock()
	if !fast {
		dbPeer, ok = peermap.peers[*id]
	}
	peermap.peers[*id] = p
	peermap.Unlock()

	if !fast {
		if ok { // Already in db
			if dbPeer.Complete == false && p.Complete == true { // They completed
				shared.AddExpval(&shared.Expvar.Leeches, -1)
				shared.AddExpval(&shared.Expvar.Seeds, 1)
			}
			if dbPeer.Complete == true && p.Complete == false { // They uncompleted?
				shared.AddExpval(&shared.Expvar.Seeds, -1)
				shared.AddExpval(&shared.Expvar.Leeches, 1)
			}
			if dbPeer.IP != p.IP { // IP changed
				shared.Expvar.IPs.Lock()
				// shared.Expvar.IPs.delete(dbPeer.IP)
				// shared.Expvar.IPs.inc(p.IP)
				shared.Expvar.IPs.Unlock()
			}
		} else { // New
			shared.Expvar.IPs.Lock()
			// shared.Expvar.IPs.inc(p.IP)
			shared.Expvar.IPs.Unlock()
			if p.Complete {
				shared.AddExpval(&shared.Expvar.Seeds, 1)
			} else {
				shared.AddExpval(&shared.Expvar.Leeches, 1)
			}
		}
	}
}

// delete is like drop but doesn't lock
func (db *Memory) delete(p *shared.Peer, h *shared.Hash, id *shared.PeerID) {
	peermap, ok := db.hashmap[*h]
	if !ok {
		return
	}
	delete(peermap.peers, *id)

	if !fast {
		if p.Complete {
			shared.AddExpval(&shared.Expvar.Seeds, -1)
		} else {
			shared.AddExpval(&shared.Expvar.Leeches, -1)
		}

		shared.Expvar.IPs.Lock()
		// shared.Expvar.IPs.dec(p.IP)
		// if shared.Expvar.IPs.dead(p.IP) {
		// 	shared.Expvar.IPs.delete(p.IP)
		// }
		shared.Expvar.IPs.Unlock()
	}
}

// Drop deletes peer
func (db *Memory) Drop(p *shared.Peer, h *shared.Hash, id *shared.PeerID) {
	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		return
	}

	peermap.Lock()
	delete(peermap.peers, *id)
	peermap.Unlock()

	if !fast {
		if p.Complete {
			shared.AddExpval(&shared.Expvar.Seeds, -1)
		} else {
			shared.AddExpval(&shared.Expvar.Leeches, -1)
		}

		shared.Expvar.IPs.Lock()
		// shared.Expvar.IPs.dec(p.IP)
		// if shared.Expvar.IPs.dead(p.IP) {
		// 	shared.Expvar.IPs.delete(p.IP)
		// }
		shared.Expvar.IPs.Unlock()
	}
}
