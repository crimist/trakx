package gomap

import (
	"github.com/crimist/trakx/tracker/storage"
)

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
				storage.Expvar.IPs.Remove(dbPeer.IP)
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

// delete is similar to drop but doesn't lock
func (db *Memory) delete(p *storage.Peer, pmap *subPeerMap, id *storage.PeerID) {
	delete(pmap.peers, *id)

	if !fast {
		if p.Complete {
			storage.AddExpval(&storage.Expvar.Seeds, -1)
		} else {
			storage.AddExpval(&storage.Expvar.Leeches, -1)
		}

		storage.Expvar.IPs.Lock()
		storage.Expvar.IPs.Remove(p.IP)
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
		storage.Expvar.IPs.Remove(p.IP)
		storage.Expvar.IPs.Unlock()
	}
}
