package shared

import (
	"sync/atomic"

	"go.uber.org/zap"
)

type PeerID [20]byte
type PeerIP [4]byte

// Peer holds peer information stores in the database
type Peer struct {
	Complete bool
	IP       PeerIP
	Port     uint16
	LastSeen int64
}

// Save creates or updates peer
func (db *PeerDatabase) Save(p *Peer, h *Hash, id *PeerID) {
	db.mu.Lock()

	if _, ok := db.db[*h]; !ok {
		db.db[*h] = make(map[PeerID]Peer)
		if !db.conf.Trakx.Prod {
			db.logger.Info("Created hash map", zap.Any("hash", h[:]))
		}
	}

	dbPeer, ok := db.db[*h][*id]
	db.db[*h][*id] = *p
	db.mu.Unlock()

	if ok { // Already in db
		if dbPeer.Complete == false && p.Complete == true { // They completed
			atomic.AddInt64(&Expvar.Leeches, -1)
			atomic.AddInt64(&Expvar.Seeds, 1)
		}
		if dbPeer.Complete == true && p.Complete == false { // They uncompleted?
			atomic.AddInt64(&Expvar.Seeds, -1)
			atomic.AddInt64(&Expvar.Leeches, 1)
		}
		if dbPeer.IP != p.IP { // IP changed
			Expvar.IPs.Lock()
			delete(Expvar.IPs.M, dbPeer.IP)
			Expvar.IPs.M[p.IP]++
			Expvar.IPs.Unlock()
		}
	} else { // New
		Expvar.IPs.Lock()
		Expvar.IPs.M[p.IP]++
		Expvar.IPs.Unlock()
		if p.Complete {
			atomic.AddInt64(&Expvar.Seeds, 1)
		} else {
			atomic.AddInt64(&Expvar.Leeches, 1)
		}
	}
}

func (db *PeerDatabase) deletePeer(p *Peer, h *Hash, id *PeerID) {
	if peer, ok := db.db[*h][*id]; ok {
		if peer.Complete {
			atomic.AddInt64(&Expvar.Seeds, -1)
		} else {
			atomic.AddInt64(&Expvar.Leeches, -1)
		}
	}
	delete(db.db[*h], *id)
}

func (db *PeerDatabase) deleteIP(ip PeerIP) {
	Expvar.IPs.M[ip]--
	if Expvar.IPs.M[ip] < 1 {
		delete(Expvar.IPs.M, ip)
	}
}

// Drop deletes peer
func (db *PeerDatabase) Drop(p *Peer, h *Hash, id *PeerID) {
	db.mu.Lock()
	db.deletePeer(p, h, id)
	db.mu.Unlock()

	Expvar.IPs.Lock()
	db.deleteIP(p.IP)
	Expvar.IPs.Unlock()
}
