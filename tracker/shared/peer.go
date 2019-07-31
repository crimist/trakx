package shared

import (
	"sync/atomic"

	"go.uber.org/zap"
)

type PeerID [20]byte
type PeerIP [4]byte

// Peer holds peer information stores in the database
type Peer struct {
	IP       PeerIP
	Port     uint16
	Complete bool
	LastSeen int64
}

// Save creates or updates peer
func (db *PeerDatabase) Save(p *Peer, h *Hash, id *PeerID) {
	db.mu.Lock()

	if _, ok := db.db[*h]; !ok {
		db.db[*h] = make(map[PeerID]Peer)
		if !Config.Trakx.Prod {
			Logger.Info("Created hash map", zap.Any("hash", h[:]))
		}
	}

	dbPeer, ok := db.db[*h][*id]
	db.db[*h][*id] = *p
	db.mu.Unlock()

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
}

func (db *PeerDatabase) deletePeer(p *Peer, h *Hash, id *PeerID) {
	if peer, ok := db.db[*h][*id]; ok {
		if peer.Complete {
			atomic.AddInt64(&ExpvarSeeds, -1)
		} else {
			atomic.AddInt64(&ExpvarLeeches, -1)
		}
	}
	delete(db.db[*h], *id)
}

func (db *PeerDatabase) deleteIP(ip PeerIP) {
	ExpvarIPs.M[ip]--
	if ExpvarIPs.M[ip] < 1 {
		delete(ExpvarIPs.M, ip)
	}
}

// Drop deletes peer
func (db *PeerDatabase) Drop(p *Peer, h *Hash, id *PeerID) {
	db.mu.Lock()
	db.deletePeer(p, h, id)
	db.mu.Unlock()

	ExpvarIPs.Lock()
	db.deleteIP(p.IP)
	ExpvarIPs.Unlock()
}
