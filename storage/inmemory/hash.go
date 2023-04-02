package inmemory

import (
	"encoding/binary"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/storage"
)

// Hashes gets the number of hashes
func (db *InMemory) Hashes() int {
	// race condition but doesn't matter as it's just for metrics
	return len(db.hashes)
}

// HashStats returns number of complete and incomplete peers associated with the hash
func (db *InMemory) HashStats(hash storage.Hash) (complete, incomplete uint16) {
	db.mutex.RLock()
	peermap, ok := db.hashes[hash]
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
func (db *InMemory) PeerList(hash storage.Hash, numWant uint, removePeerId bool) (peers [][]byte) {
	db.mutex.RLock()
	peermap, ok := db.hashes[hash]
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
	dictionary := pools.Dictionaries.Get()

	for id, peer := range peermap.Peers {
		if !removePeerId {
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

	peermap.mutex.RUnlock()
	pools.Dictionaries.Put(dictionary)

	return
}

// PeerListBytes returns a byte encoded peer list for the given hash capped at num
func (db *InMemory) PeerListBytes(hash storage.Hash, numWant uint) (peers4 []byte, peers6 []byte) {
	peers4 = pools.Peerlists4.Get()
	peers6 = pools.Peerlists6.Get()

	db.mutex.RLock()
	peermap, ok := db.hashes[hash]
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
			copy(peers6[pos6:pos6+16], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers6[pos6+16:pos6+18], peer.Port)
			pos6 += 18
			if pos6+18 > cap(peers6) {
				break
			}
		} else {
			copy(peers4[pos4:pos4+4], peer.IP.AsSlice())
			binary.BigEndian.PutUint16(peers4[pos4+4:pos4+6], peer.Port)
			pos4 += 6
			if pos4+6 > cap(peers4) {
				break
			}
		}
	}
	peermap.mutex.RUnlock()

	peers4 = peers4[:pos4]
	peers6 = peers6[:pos6]

	return
}
