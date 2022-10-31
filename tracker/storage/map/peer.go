package gomap

import (
	"net/netip"
	"time"

	"github.com/crimist/trakx/tracker/storage"
)

func (memoryDb *Memory) Save(ip netip.Addr, port uint16, complete bool, hash storage.Hash, id storage.PeerID) {
	// get/create the map
	memoryDb.mutex.RLock()
	peermap, ok := memoryDb.hashmap[hash]
	memoryDb.mutex.RUnlock()

	// if submap doesn't exist create it
	if !ok {
		memoryDb.mutex.Lock()
		peermap = memoryDb.makePeermap(hash)
		memoryDb.mutex.Unlock()
	}

	// get peer
	peermap.mutex.RLock()
	peer, peerExists := peermap.Peers[id]
	peermap.mutex.RUnlock()

	peermap.mutex.Lock()
	// if peer does not exist then create
	if !peerExists {
		peer = storage.PeerChan.Get()
		peermap.Peers[id] = peer
	}

	// update peermap completion counts
	// TODO: consider using atomic package instead of locking peermap?
	if peerExists {
		if !peer.Complete && complete {
			peermap.Incomplete--
			peermap.Complete++
		} else if peer.Complete && !complete {
			peermap.Complete--
			peermap.Incomplete++
		}
	} else {
		if complete {
			peermap.Complete++
		} else {
			peermap.Incomplete++
		}
	}
	peermap.mutex.Unlock()

	// update metrics
	if !fast {
		if peerExists {
			// They completed
			if !peer.Complete && complete {
				storage.Expvar.Leeches.Add(-1)
				storage.Expvar.Seeds.Add(1)
			} else if peer.Complete && !complete { // They uncompleted?
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
func (db *Memory) delete(peer *storage.Peer, peermap *PeerMap, id storage.PeerID) {
	delete(peermap.Peers, id)

	if peer.Complete {
		peermap.Complete--
	} else {
		peermap.Incomplete--
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
func (db *Memory) Drop(hash storage.Hash, id storage.PeerID) {
	// get the peermap
	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	// get the peer and remove it
	peermap.mutex.Lock()
	peer, ok := peermap.Peers[id]
	if !ok {
		peermap.mutex.Unlock()
		return
	}
	delete(peermap.Peers, id)

	if peer.Complete {
		peermap.Complete--
	} else {
		peermap.Incomplete--
	}
	peermap.mutex.Unlock()

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
