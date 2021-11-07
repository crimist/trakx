package gomap

import (
	"time"

	"github.com/crimist/trakx/tracker/storage"
)

func (db *Memory) Save(ip storage.PeerIP, port uint16, complete bool, h storage.Hash, id storage.PeerID) {
	// get/create the map
	db.mu.RLock()
	peermap, mapExists := db.hashmap[h]
	db.mu.RUnlock()

	// if submap doesn't exist create it
	if !mapExists {
		db.mu.Lock()
		peermap = db.makePeermap(h)
		db.mu.Unlock()
	}

	// get peer
	peermap.RLock()
	peer, peerExists := peermap.peers[id]
	peermap.RUnlock()

	// if peer does not exist then create
	if !peerExists {
		peermap.Lock()
		peer = storage.PeerChan.Get()
		peermap.peers[id] = peer
		peermap.Unlock()
	}

	// update peermap completion counts
	if peerExists {
		if peer.Complete == false && complete == true {
			peermap.incomplete--
			peermap.complete++
		} else if peer.Complete == true && complete == false {
			peermap.complete--
			peermap.incomplete++
		}
	} else {
		if complete == true {
			peermap.complete++
		} else {
			peermap.incomplete++
		}
	}

	// update metrics
	if !fast {
		if peerExists {
			// They completed
			if peer.Complete == false && complete == true {
				storage.Expvar.Leeches.Add(-1)
				storage.Expvar.Seeds.Add(1)
			} else if peer.Complete == true && complete == false { // They uncompleted?
				storage.Expvar.Seeds.Add(-1)
				storage.Expvar.Leeches.Add(1)
			}
			// IP changed
			if peer.IP != ip {
				storage.Expvar.IPs.Lock()
				storage.Expvar.IPs.Remove(peer.IP)
				storage.Expvar.IPs.Inc(ip)
				storage.Expvar.IPs.Unlock()
			}
		} else {
			storage.Expvar.IPs.Lock()
			storage.Expvar.IPs.Inc(ip)
			storage.Expvar.IPs.Unlock()

			if complete {
				storage.Expvar.Seeds.Add(1)
			} else {
				storage.Expvar.Leeches.Add(1)
			}
		}
	}

	// update peer
	peer.Complete = complete
	peer.IP = ip
	peer.Port = port
	peer.LastSeen = time.Now().Unix()
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
