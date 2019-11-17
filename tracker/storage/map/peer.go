package gomap

import (
	"github.com/crimist/trakx/tracker/storage"
)

func (db *Memory) makePeermap(h *storage.Hash) (peermap *subPeerMap) {
	// build struct and assign
	peermap = new(subPeerMap)
	peermap.peers = make(map[storage.PeerID]*storage.Peer, 1)
	db.hashmap[*h] = peermap
	return
}

// Save writes a peer
func (db *Memory) Save(p *storage.Peer, h *storage.Hash, id *storage.PeerID) {
	var dbPeer *storage.Peer

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
				storage.AddExpval(&storage.Expvar.Leeches, -1)
				storage.AddExpval(&storage.Expvar.Seeds, 1)
			} else if dbPeer.Complete == true && p.Complete == false { // They uncompleted?
				storage.AddExpval(&storage.Expvar.Seeds, -1)
				storage.AddExpval(&storage.Expvar.Leeches, 1)
			}
			if dbPeer.IP != p.IP { // IP changed
				storage.Expvar.IPs.Lock()
				storage.Expvar.IPs.Delete(dbPeer.IP)
				storage.Expvar.IPs.Inc(p.IP)
				storage.Expvar.IPs.Unlock()
			}
		} else { // New
			storage.Expvar.IPs.Lock()
			storage.Expvar.IPs.Inc(p.IP)
			storage.Expvar.IPs.Unlock()
			if p.Complete {
				storage.AddExpval(&storage.Expvar.Seeds, 1)
			} else {
				storage.AddExpval(&storage.Expvar.Leeches, 1)
			}
		}
	}
}

// delete is like drop but doesn't lock
func (db *Memory) delete(p *storage.Peer, h *storage.Hash, id *storage.PeerID) {
	peermap, ok := db.hashmap[*h]
	if !ok {
		return
	}
	delete(peermap.peers, *id)

	if !fast {
		if p.Complete {
			storage.AddExpval(&storage.Expvar.Seeds, -1)
		} else {
			storage.AddExpval(&storage.Expvar.Leeches, -1)
		}

		storage.Expvar.IPs.Lock()
		storage.Expvar.IPs.Dec(p.IP)
		if storage.Expvar.IPs.Dead(p.IP) {
			storage.Expvar.IPs.Delete(p.IP)
		}
		storage.Expvar.IPs.Unlock()
	}
}

// Drop deletes peer
func (db *Memory) Drop(p *storage.Peer, h *storage.Hash, id *storage.PeerID) {
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
			storage.AddExpval(&storage.Expvar.Seeds, -1)
		} else {
			storage.AddExpval(&storage.Expvar.Leeches, -1)
		}

		storage.Expvar.IPs.Lock()
		storage.Expvar.IPs.Dec(p.IP)
		if storage.Expvar.IPs.Dead(p.IP) {
			storage.Expvar.IPs.Delete(p.IP)
		}
		storage.Expvar.IPs.Unlock()
	}
}
