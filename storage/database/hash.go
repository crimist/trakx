package database

import (
	"encoding/binary"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/storage"
)

func (db *Database) TorrentStats(hash storage.Hash) (seeds, leeches uint16) {
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

func (db *Database) TorrentPeers(hash storage.Hash, numWant uint, includePeerID bool) (peers [][]byte) {
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
	dictionary := bencoding.AcquireDictionary()

	torrent.mutex.RLock()
	for id, peer := range torrent.Peers {
		if includePeerID {
			dictionary.String("peer id", string(id[:]))
		}
		dictionary.String("ip", peer.IP.String())
		dictionary.Int64("port", int64(peer.Port))

		dictBytes := dictionary.GetBytes()
		peers[i] = make([]byte, len(dictBytes))
		copy(peers[i], dictBytes)
		dictionary.Reset()

		i++
		if i == numWant {
			break
		}
	}
	torrent.mutex.RUnlock()

	bencoding.ReleaseDictionary(dictionary)
	return
}

func (db *Database) TorrentPeersCompact(hash storage.Hash, numWant uint, wantedIPs storage.IPVersion) storage.PeerLists {
	db.mutex.RLock()
	torrent, ok := db.torrents[hash]
	db.mutex.RUnlock()
	if !ok {
		return storage.NewPeerLists(nil, nil, nil)
	}

	torrent.mutex.RLock()
	numPeers := uint(len(torrent.Peers))
	torrent.mutex.RUnlock()

	if numWant > numPeers {
		numWant = numPeers
	}
	if numWant == 0 {
		return storage.NewPeerLists(nil, nil, nil)
	}
	if wantedIPs == 0 {
		return storage.NewPeerLists(nil, nil, nil)
	}

	var peers4, peers6 []byte
	if wantedIPs&storage.IPv4 != 0 {
		if db.peerLists != nil {
			peers4 = db.peerLists.getV4()
		} else {
			peers4 = make([]byte, int(numWant)*peerlist4Stride)
		}
	}
	if wantedIPs&storage.IPv6 != 0 {
		if db.peerLists != nil {
			peers6 = db.peerLists.getV6()
		} else {
			peers6 = make([]byte, int(numWant)*peerlist6Stride)
		}
	}

	var pos4, pos6 int
	max4Bytes := int(numWant) * peerlist4Stride
	max6Bytes := int(numWant) * peerlist6Stride
	if peers4 != nil && max4Bytes > len(peers4) {
		max4Bytes = len(peers4)
	}
	if peers6 != nil && max6Bytes > len(peers6) {
		max6Bytes = len(peers6)
	}

	torrent.mutex.RLock()
	for _, peer := range torrent.Peers {
		if peer.IP.Is6() {
			if wantedIPs&storage.IPv6 == 0 {
				continue
			}
			if pos6+peerlist6Stride > max6Bytes {
				break
			}
			copy(peers6[pos6:pos6+16], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers6[pos6+16:pos6+18], peer.Port)
			pos6 += peerlist6Stride
			if pos6+peerlist6Stride > max6Bytes {
				break
			}
			continue
		}
		if wantedIPs&storage.IPv4 == 0 {
			continue
		}
		if pos4+peerlist4Stride > max4Bytes {
			break
		}
		copy(peers4[pos4:pos4+4], peer.IP.AsSlice())
		binary.BigEndian.PutUint16(peers4[pos4+4:pos4+6], peer.Port)
		pos4 += peerlist4Stride
		if pos4+peerlist4Stride > max4Bytes {
			break
		}
	}
	torrent.mutex.RUnlock()

	if wantedIPs&storage.IPv4 != 0 {
		peers4 = peers4[:pos4]
	}
	if wantedIPs&storage.IPv6 != 0 {
		peers6 = peers6[:pos6]
	}

	release := func() {
		if db.peerLists != nil {
			db.peerLists.putV4(peers4)
			db.peerLists.putV6(peers6)
		}
	}

	return storage.NewPeerLists(peers4, peers6, release)
}
