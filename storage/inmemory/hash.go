package inmemory

import (
	"encoding/binary"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/utils"
)

func (db *InMemory) TorrentStats(hash storage.Hash) (seeds, leeches uint16) {
	db.mutex.RLock()
	torrent, ok := db.torrents[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	torrent.mutex.RLock()
	seeds = torrent.Seeds
	leeches = torrent.Leeches
	torrent.mutex.RUnlock()

	return
}

func (db *InMemory) TorrentPeers(hash storage.Hash, numWant uint, includePeerID bool) (peers [][]byte) {
	db.mutex.RLock()
	torrent, ok := db.torrents[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	// TODO: benchmark the performance of mutex placement
	torrent.mutex.RLock()
	numPeers := uint(len(torrent.Peers))
	torrent.mutex.RUnlock()

	if numWant > numPeers {
		numWant = numPeers
	}
	if numWant == 0 {
		return
	}

	var i uint
	peers = make([][]byte, numWant)
	dictionary := pools.Dictionaries.Get()

	torrent.mutex.RLock()
	for id, peer := range torrent.Peers {
		if includePeerID {
			dictionary.String("peer id", utils.ByteToStringUnsafe(id[:]))
		}
		dictionary.String("ip", peer.IP.String())
		dictionary.Int64("port", int64(peer.Port))

		// peers[i] = make([]byte, len(data))
		// copy(peers[i], data)
		// TODO: test this
		peers[i] = dictionary.GetBytes()

		dictionary.Reset()

		i++
		if i == numWant {
			break
		}
	}
	torrent.mutex.RUnlock()

	pools.Dictionaries.Put(dictionary)
	return
}

func (db *InMemory) TorrentPeersCompact(hash storage.Hash, numWant uint, wantedIPs storage.IPVersion) (peers4 []byte, peers6 []byte) {
	if wantedIPs&storage.IPv4 != 0 {
		peers4 = pools.Peerlists4.Get()
	}
	if wantedIPs&storage.IPv6 != 0 {
		peers6 = pools.Peerlists6.Get()
	}

	db.mutex.RLock()
	torrent, ok := db.torrents[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	torrent.mutex.RLock()
	numPeers := uint(len(torrent.Peers))
	if numWant > numPeers {
		numWant = numPeers
	}
	if numWant == 0 {
		torrent.mutex.RUnlock()
		return
	}

	var pos4, pos6 int
	for _, peer := range torrent.Peers {
		if peer.IP.Is6() && wantedIPs&storage.IPv6 != 0 {
			copy(peers6[pos6:pos6+16], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers6[pos6+16:pos6+18], peer.Port)
			pos6 += 18
			if pos6+18 > cap(peers6) {
				break
			}
		} else if wantedIPs&storage.IPv4 != 0 {
			copy(peers4[pos4:pos4+4], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers4[pos4+4:pos4+6], peer.Port)
			pos4 += 6
			if pos4+6 > cap(peers4) {
				break
			}
		}
	}
	torrent.mutex.RUnlock()

	if wantedIPs&storage.IPv4 != 0 {
		peers4 = peers4[:pos4]
	}
	if wantedIPs&storage.IPv6 != 0 {
		peers6 = peers6[:pos6]
	}

	return
}
