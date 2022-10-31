package gomap

import (
	"encoding/binary"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/storage"
)

// Hashes gets the number of hashes
func (db *Memory) Hashes() int {
	// race condition but doesn't matter as it's just for metrics
	return len(db.hashmap)
}

// HashStats returns number of complete and incomplete peers associated with the hash
func (db *Memory) HashStats(hash storage.Hash) (complete, incomplete uint16) {
	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	peermap.mutex.RLock()
	complete = peermap.Complete
	incomplete = peermap.Incomplete
	peermap.mutex.RUnlock()

	return
}

// PeerList returns a peer list for the given hash capped at max
func (db *Memory) PeerList(hash storage.Hash, numWant uint, removePeerId bool) (peers [][]byte) {
	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	peermap.mutex.RLock()

	if numPeers := uint(len(peermap.Peers)); numWant > numPeers {
		numWant = numPeers
	}

	if numWant == 0 {
		peermap.mutex.RUnlock()
		return
	}

	var i uint
	peers = make([][]byte, numWant)
	dict := bencoding.GetDictionary()

	for id, peer := range peermap.Peers {
		if !removePeerId {
			dict.String("peer id", string(id[:]))
		}
		dict.String("ip", peer.IP.String())
		dict.Int64("port", int64(peer.Port))

		dictBytes := dict.GetBytes()
		peers[i] = make([]byte, len(dictBytes))
		copy(peers[i], dictBytes)

		dict.Reset()

		i++
		if i == numWant {
			break
		}
	}

	peermap.mutex.RUnlock()
	bencoding.PutDictionary(dict)

	return
}

// PeerListBytes returns a byte encoded peer list for the given hash capped at num
func (db *Memory) PeerListBytes(hash storage.Hash, numWant uint) (peers4 *storage.Peerlist, peers6 *storage.Peerlist) {
	peers4 = storage.GetPeerList()
	peers6 = storage.GetPeerList()

	db.mutex.RLock()
	peermap, ok := db.hashmap[hash]
	db.mutex.RUnlock()
	if !ok {
		return
	}

	peermap.mutex.RLock()
	if numPeers := uint(len(peermap.Peers)); numWant > numPeers {
		numWant = numPeers
	}

	if numWant == 0 {
		peermap.mutex.RUnlock()
		return
	}

	var pos4, pos6 int
	for _, peer := range peermap.Peers {
		if peer.IP.Is6() {
			copy(peers6.Data[pos6:pos6+16], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers6.Data[pos6+16:pos6+18], peer.Port)
			pos6 += 18
			if pos6+18 > cap(peers6.Data) {
				break
			}
		} else {
			copy(peers4.Data[pos4:pos4+4], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers4.Data[pos4+4:pos4+6], peer.Port)
			pos4 += 6
			if pos4+6 > cap(peers4.Data) {
				break
			}
		}
	}
	peermap.mutex.RUnlock()

	peers4.Data = peers4.Data[:pos4]
	peers6.Data = peers6.Data[:pos6]

	return
}
