package gomap

import (
	"github.com/crimist/trakx/tracker/storage"
)

// Save writes a peer
func (db *Memory) Save(peer *storage.Peer, h *storage.Hash, id *storage.PeerID) {
	// get/create the map
	db.mu.RLock()
	peermap, mapExists := db.hashmap[*h]
	db.mu.RUnlock()
	if !mapExists {
		db.mu.Lock()
		peermap = db.makePeermap(h)
		db.mu.Unlock()
	}

	// assign the peer
	peermap.Lock()
	oldpeer, peerExists := peermap.peers[*id]
	peermap.peers[*id] = peer
	peermap.Unlock()

	if !fast {
		// metric calculation

		if peerExists {
			// They completed
			if oldpeer.Complete == false && peer.Complete == true {
				storage.AddExpval(&storage.Expvar.Leeches, -1)
				storage.AddExpval(&storage.Expvar.Seeds, 1)
			} else if oldpeer.Complete == true && peer.Complete == false { // They uncompleted?
				storage.AddExpval(&storage.Expvar.Seeds, -1)
				storage.AddExpval(&storage.Expvar.Leeches, 1)
			}
			// IP changed
			if oldpeer.IP != peer.IP {
				storage.Expvar.IPs.Lock()
				storage.Expvar.IPs.Remove(oldpeer.IP)
				storage.Expvar.IPs.Inc(peer.IP)
				storage.Expvar.IPs.Unlock()
			}
		} else {
			storage.Expvar.IPs.Lock()
			storage.Expvar.IPs.Inc(peer.IP)
			storage.Expvar.IPs.Unlock()

			if peer.Complete {
				storage.AddExpval(&storage.Expvar.Seeds, 1)
			} else {
				storage.AddExpval(&storage.Expvar.Leeches, 1)
			}
		}
	}

	// put back the old peer if it exists
	if peerExists {
		storage.PutPeer(oldpeer)
	}
}

// delete is similar to drop but doesn't lock
func (db *Memory) delete(peer *storage.Peer, pmap *PeerMap, id *storage.PeerID) {
	delete(pmap.peers, *id)

	if !fast {
		if peer.Complete {
			storage.AddExpval(&storage.Expvar.Seeds, -1)
		} else {
			storage.AddExpval(&storage.Expvar.Leeches, -1)
		}

		storage.Expvar.IPs.Lock()
		storage.Expvar.IPs.Remove(peer.IP)
		storage.Expvar.IPs.Unlock()
	}

	storage.PutPeer(peer)
}

// Drop deletes peer
func (db *Memory) Drop(h *storage.Hash, id *storage.PeerID) {
	var peer *storage.Peer

	// get the peermap
	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		return
	}

	// get the peer and remove it
	peermap.Lock()
	peer, ok = peermap.peers[*id]
	if !ok {
		peermap.Unlock()
		return
	}
	delete(peermap.peers, *id)
	peermap.Unlock()

	if !fast {
		if peer.Complete {
			storage.AddExpval(&storage.Expvar.Seeds, -1)
		} else {
			storage.AddExpval(&storage.Expvar.Leeches, -1)
		}

		storage.Expvar.IPs.Lock()
		storage.Expvar.IPs.Remove(peer.IP)
		storage.Expvar.IPs.Unlock()
	}

	// free the peer back to the pool
	storage.PutPeer(peer)
}
