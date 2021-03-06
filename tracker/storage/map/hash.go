package gomap

import (
	"encoding/binary"
	"net"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/storage"
)

// Hashes gets the number of hashes
func (db *Memory) Hashes() int {
	// race condition but doesn't matter as it's just for metrics
	return len(db.hashmap)
}

// HashStats returns number of complete and incomplete peers associated with the hash
func (db *Memory) HashStats(h storage.Hash) (complete, incomplete uint16) {
	db.mu.RLock()
	peermap, ok := db.hashmap[h]
	db.mu.RUnlock()
	if !ok {
		return
	}

	peermap.RLock()
	complete = peermap.complete
	incomplete = peermap.incomplete
	peermap.RUnlock()

	return
}

// PeerList returns a peer list for the given hash capped at max
func (db *Memory) PeerList(h storage.Hash, max int, noPeerID bool) [][]byte {
	db.mu.RLock()
	peermap, ok := db.hashmap[h]
	db.mu.RUnlock()
	if !ok {
		return [][]byte{}
	}

	peermap.RLock()

	if mlen := len(peermap.peers); max > mlen {
		max = mlen
	}

	if max == 0 {
		peermap.RUnlock()
		return [][]byte{}
	}

	var i int
	peerList := make([][]byte, max)
	dict := bencoding.GetDictionary()

	for id, peer := range peermap.peers {
		if !noPeerID {
			dict.String("peer id", string(id[:]))
		}
		dict.String("ip", net.IP(peer.IP[:]).String())
		dict.Int64("port", int64(peer.Port))

		b := dict.GetBytes()
		peerList[i] = make([]byte, len(b))
		copy(peerList[i], b)

		dict.Reset()

		i++
		if i == max {
			break
		}
	}

	peermap.RUnlock()
	bencoding.PutDictionary(dict)

	return peerList
}

// PeerListBytes returns a byte encoded peer list for the given hash capped at num
func (db *Memory) PeerListBytes(h storage.Hash, max int) *storage.Peerlist {
	plist := storage.GetPeerList()

	db.mu.RLock()
	peermap, ok := db.hashmap[h]
	db.mu.RUnlock()
	if !ok {
		return plist
	}

	var pos int

	peermap.RLock()
	if mlen := len(peermap.peers); max > mlen {
		max = mlen
	}

	if max == 0 {
		peermap.RUnlock()
		return plist
	}

	size := 6 * max
	plist.Data = plist.Data[:size]
	for _, peer := range peermap.peers {
		copy(plist.Data[pos:pos+4], peer.IP[:])
		binary.BigEndian.PutUint16(plist.Data[pos+4:pos+6], peer.Port)

		pos += 6
		if pos == size {
			break
		}
	}
	peermap.RUnlock()

	return plist
}
