package inmemory

import (
	"github.com/syc0x00/trakx/tracker/database"
	"github.com/syc0x00/trakx/tracker/shared"
)

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
				database.AddExpval(&database.Expvar.Leeches, -1)
				database.AddExpval(&database.Expvar.Seeds, 1)
			} else if dbPeer.Complete == true && p.Complete == false { // They uncompleted?
				database.AddExpval(&database.Expvar.Seeds, -1)
				database.AddExpval(&database.Expvar.Leeches, 1)
			}
			if dbPeer.IP != p.IP { // IP changed
				database.Expvar.IPs.Lock()
				database.Expvar.IPs.Delete(dbPeer.IP)
				database.Expvar.IPs.Inc(p.IP)
				database.Expvar.IPs.Unlock()
			}
		} else { // New
			database.Expvar.IPs.Lock()
			database.Expvar.IPs.Inc(p.IP)
			database.Expvar.IPs.Unlock()
			if p.Complete {
				database.AddExpval(&database.Expvar.Seeds, 1)
			} else {
				database.AddExpval(&database.Expvar.Leeches, 1)
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
			database.AddExpval(&database.Expvar.Seeds, -1)
		} else {
			database.AddExpval(&database.Expvar.Leeches, -1)
		}

		database.Expvar.IPs.Lock()
		database.Expvar.IPs.Dec(p.IP)
		if database.Expvar.IPs.Dead(p.IP) {
			database.Expvar.IPs.Delete(p.IP)
		}
		database.Expvar.IPs.Unlock()
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
			database.AddExpval(&database.Expvar.Seeds, -1)
		} else {
			database.AddExpval(&database.Expvar.Leeches, -1)
		}

		database.Expvar.IPs.Lock()
		database.Expvar.IPs.Dec(p.IP)
		if database.Expvar.IPs.Dead(p.IP) {
			database.Expvar.IPs.Delete(p.IP)
		}
		database.Expvar.IPs.Unlock()
	}
}
