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
	if peerExists {
		if !peer.Complete && complete {
			db.collector.AddLeeches(-1)
			db.collector.AddSeeds(1)
		} else if peer.Complete && !complete {
			db.collector.AddSeeds(-1)
			db.collector.AddLeeches(1)
		}

		db.collector.IPs().Replace(peer.IP, ip)
	} else {
		db.collector.IPs().Inc(ip)

		if complete {
			db.collector.AddSeeds(1)
		} else {
			db.collector.AddLeeches(1)
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

	if peer.Complete {
		db.collector.AddSeeds(-1)
	} else {
		db.collector.AddLeeches(-1)
	}
	db.collector.IPs().Remove(peer.IP)

	db.peerPool.Put(peer)
}
