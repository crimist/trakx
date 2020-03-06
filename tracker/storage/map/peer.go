package gomap

import (
	"github.com/crimist/trakx/tracker/storage"
)

// Save writes a peer
func (db *Memory) Save(peer *storage.Peer, h storage.Hash, id storage.PeerID) {
	// get/create the map
	db.mu.RLock()
	peermap, mapExists := db.hashmap[h]
	db.mu.RUnlock()

	if !mapExists {
		db.mu.Lock()
		peermap = db.makePeermap(h)
		db.mu.Unlock()
	}

	// assign the peer
	peermap.Lock()
	oldpeer, peerExists := peermap.peers[id]
	peermap.peers[id] = peer

	if peerExists {
		if oldpeer.Complete == false && peer.Complete == true {
			peermap.incomplete--
			peermap.complete++
		} else if oldpeer.Complete == true && peer.Complete == false {
			peermap.complete--
			peermap.incomplete++
		}
	} else {
		if peer.Complete == true {
			peermap.complete++
		} else {
			peermap.incomplete++
		}
	}

	peermap.Unlock()

	if !fast {
		// metric calculation
		if peerExists {
			// They completed
			if oldpeer.Complete == false && peer.Complete == true {
				storage.Expvar.Leeches.Add(-1)
				storage.Expvar.Seeds.Add(1)
			} else if oldpeer.Complete == true && peer.Complete == false { // They uncompleted?
				storage.Expvar.Seeds.Add(-1)
				storage.Expvar.Leeches.Add(1)
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
				storage.Expvar.Seeds.Add(1)
			} else {
				storage.Expvar.Leeches.Add(1)
			}
		}
	}

	// put back the old peer if it exists
	if peerExists {
		storage.PeerChan.Put(oldpeer)
	}
}

// delete is similar to drop but doesn't lock
func (db *Memory) delete(peer *storage.Peer, pmap *PeerMap, id storage.PeerID) {
	delete(pmap.peers, id)

	if peer.Complete == true {
		pmap.complete--
	} else {
		pmap.incomplete--
	}

	if !fast {
		if peer.Complete {
			storage.Expvar.Seeds.Add(-1)
		} else {
			storage.Expvar.Leeches.Add(-1)
		}

		storage.Expvar.IPs.Lock()
		storage.Expvar.IPs.Remove(peer.IP)
		storage.Expvar.IPs.Unlock()
	}

	storage.PeerChan.Put(peer)
}

// Drop deletes peer
func (db *Memory) Drop(h storage.Hash, id storage.PeerID) {
	// get the peermap
	db.mu.RLock()
	peermap, ok := db.hashmap[h]
	db.mu.RUnlock()
	if !ok {
		return
	}

	// get the peer and remove it
	peermap.Lock()
	peer, ok := peermap.peers[id]
	if !ok {
		peermap.Unlock()
		return
	}
	delete(peermap.peers, id)

	if peer.Complete == true {
		peermap.complete--
	} else {
		peermap.incomplete--
	}
	peermap.Unlock()

	if !fast {
		if peer.Complete {
			storage.Expvar.Seeds.Add(-1)
		} else {
			storage.Expvar.Leeches.Add(-1)
		}

		storage.Expvar.IPs.Lock()
		storage.Expvar.IPs.Remove(peer.IP)
		storage.Expvar.IPs.Unlock()
	}

	// free the peer back to the pool
	storage.PeerChan.Put(peer)
}
