package database

import (
	"net/netip"
	"time"

	"github.com/crimist/trakx/internal/storage"
)

func (db *Database) PeerAdd(hash storage.Hash, id storage.PeerID, ip netip.Addr, port uint16, complete bool) {
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

	// Benchmarks run on January 10, 2026 showed atomics are faster in isolation, but since we
	// already take this lock for map updates, keeping the counter update under the same lock
	// is slightly faster than adding separate atomic ops.
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
func (db *Database) PeerRemove(hash storage.Hash, id storage.PeerID) {
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
