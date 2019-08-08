package shared

import (
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

	if expvarOn {
		if ok { // Already in db
			if dbPeer.Complete == false && p.Complete == true { // They completed
				AddExpval(&Expvar.Leeches, -1)
				AddExpval(&Expvar.Seeds, 1)
			}
			if dbPeer.Complete == true && p.Complete == false { // They uncompleted?
				AddExpval(&Expvar.Seeds, -1)
				AddExpval(&Expvar.Leeches, 1)
			}
			if dbPeer.IP != p.IP { // IP changed
				Expvar.IPs.Lock()
				Expvar.IPs.delete(dbPeer.IP)
				Expvar.IPs.inc(p.IP)
				Expvar.IPs.Unlock()
			}
		} else { // New
			Expvar.IPs.Lock()
			Expvar.IPs.inc(p.IP)
			Expvar.IPs.Unlock()
			if p.Complete {
				AddExpval(&Expvar.Seeds, 1)
			} else {
				AddExpval(&Expvar.Leeches, 1)
			}
		}
	}
}

func (db *PeerDatabase) deletePeer(p *Peer, h *Hash, id *PeerID) {
	if expvarOn {
		if peer, ok := db.db[*h][*id]; ok {
			if peer.Complete {
				AddExpval(&Expvar.Seeds, -1)
			} else {
				AddExpval(&Expvar.Leeches, -1)
			}
		}
	}

	delete(db.db[*h], *id)
}

func (db *PeerDatabase) deleteIP(ip PeerIP) {
	Expvar.IPs.dec(ip)
	if Expvar.IPs.dead(ip) {
		Expvar.IPs.delete(ip)
	}
}

// Drop deletes peer
func (db *PeerDatabase) Drop(p *Peer, h *Hash, id *PeerID) {
	db.mu.Lock()
	db.deletePeer(p, h, id)
	db.mu.Unlock()

	if expvarOn {
		Expvar.IPs.Lock()
		db.deleteIP(p.IP)
		Expvar.IPs.Unlock()
	}
}
