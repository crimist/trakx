package gomap

import (
	"github.com/crimist/trakx/tracker/storage"
)

// Save writes a peer
func (db *Memory) Save(peer *storage.Peer, h storage.Hash, id storage.PeerID) {
	var oldpeer *storage.Peer
	var peerExists bool

	db.mu.RLock()
	peermap, exists := db.hashmap[h]
	db.mu.RUnlock()
	if !exists {
		db.mu.Lock()
		peermap = db.makePeermap(h)
		db.mu.Unlock()
	}

	peermap.Lock()
	if !fast {
		oldpeer, peerExists = peermap.peers[id]
	}
	peermap.peers[id] = peer
	peermap.Unlock()

	if !fast {
		if peerExists {
			if oldpeer.Complete == false && peer.Complete == true { // They completed
				storage.AddExpval(&storage.Expvar.Leeches, -1)
				storage.AddExpval(&storage.Expvar.Seeds, 1)
			} else if oldpeer.Complete == true && peer.Complete == false { // They uncompleted?
				storage.AddExpval(&storage.Expvar.Seeds, -1)
				storage.AddExpval(&storage.Expvar.Leeches, 1)
			}
			if oldpeer.IP != peer.IP { // IP changed
				storage.Expvar.IPs.Lock()
				storage.Expvar.IPs.Remove(oldpeer.IP)
				storage.Expvar.IPs.Inc(peer.IP)
				storage.Expvar.IPs.Unlock()
			}
		} else { // New
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
}

// delete is similar to drop but doesn't lock
func (db *Memory) delete(peer *storage.Peer, pmap *subPeerMap, id storage.PeerID) {
	delete(pmap.peers, id)

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
}

// Drop deletes peer
func (db *Memory) Drop(h storage.Hash, id storage.PeerID) {
	var peer *storage.Peer

	db.mu.RLock()
	peermap, ok := db.hashmap[h]
	db.mu.RUnlock()
	if !ok {
		return
	}

	peermap.Lock()
	if !fast {
		peer, ok = peermap.peers[id]
		if !ok {
			peermap.Unlock()
			return
		}
	}
	delete(peermap.peers, id)
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
}
