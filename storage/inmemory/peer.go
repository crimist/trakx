package inmemory

import (
	"net/netip"
	"time"

	"github.com/crimist/trakx/storage"
)

func (db *InMemory) PeerAdd(hash storage.Hash, id storage.PeerID, ip netip.Addr, port uint16, complete bool) {
	db.mutex.RLock()
	torrent, torrentExists := db.torrents[hash]
	db.mutex.RUnlock()
	if !torrentExists {
		torrent = db.createTorrent(hash)
	}

	torrent.mutex.RLock()
	peer, peerExists := torrent.Peers[id]
	torrent.mutex.RUnlock()

	if !peerExists {
		peer = db.peerPool.Get()
		torrent.mutex.Lock()
		torrent.Peers[id] = peer
		torrent.mutex.Unlock()
	}

	// TODO: test if this claim of performance is true
	// raw increment is 19x faster than atomic so we might as well just wrap it in the mutex
	torrent.mutex.Lock()
	if peerExists {
		if !peer.Complete && complete {
			torrent.Leeches--
			torrent.Seeds++
		} else if peer.Complete && !complete {
			torrent.Seeds--
			torrent.Leeches++
		}
	} else {
		if complete {
			torrent.Seeds++
		} else {
			torrent.Leeches++
		}
	}
	torrent.mutex.Unlock()

	// update metrics
	if dbStats && db.stats != nil {
		if peerExists {
			if !peer.Complete && complete {
				db.stats.Leeches.Add(-1)
				db.stats.Seeds.Add(1)
			} else if peer.Complete && !complete {
				db.stats.Seeds.Add(-1)
				db.stats.Leeches.Add(1)
			}
			if peer.IP != ip {
				db.stats.IPStats.Lock()
				db.stats.IPStats.Remove(peer.IP)
				db.stats.IPStats.Inc(ip)
				db.stats.IPStats.Unlock()
			}
		} else {
			db.stats.IPStats.Lock()
			db.stats.IPStats.Inc(ip)
			db.stats.IPStats.Unlock()

			if complete {
				db.stats.Seeds.Add(1)
			} else {
				db.stats.Leeches.Add(1)
			}
		}
	}

	peer.Complete = complete
	peer.IP = ip
	peer.Port = port
	peer.LastSeen = time.Now().Unix()
}

// PeerRemove removes the given peer with id from the torrent with hash
func (db *InMemory) PeerRemove(hash storage.Hash, id storage.PeerID) {
	db.mutex.RLock()
	torrent, torrentExists := db.torrents[hash]
	db.mutex.RUnlock()
	if !torrentExists {
		return
	}

	torrent.mutex.RLock()
	peer, peerExists := torrent.Peers[id]
	torrent.mutex.RUnlock()
	if !peerExists {
		return
	}

	torrent.mutex.Lock()
	delete(torrent.Peers, id)
	if peer.Complete {
		torrent.Seeds--
	} else {
		torrent.Leeches--
	}
	torrent.mutex.Unlock()

	if dbStats && db.stats != nil {
		if peer.Complete {
			db.stats.Seeds.Add(-1)
		} else {
			db.stats.Leeches.Add(-1)
		}

		db.stats.IPStats.Lock()
		db.stats.IPStats.Remove(peer.IP)
		db.stats.IPStats.Unlock()
	}

	db.peerPool.Put(peer)
}
