package inmemory

import (
	"encoding/binary"
	"net"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
)

// Hashes gets the number of hashes
func (db *Memory) Hashes() int {
	return len(db.hashmap)
}

// HashStats returns number of complete and incomplete peers associated with the hash
func (db *Memory) HashStats(h *shared.Hash) (complete, incomplete int32) {
	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		return
	}

	peermap.RLock()
	for _, peer := range peermap.peers {
		if peer.Complete {
			complete++
		}
	}
	peermap.RUnlock()
	incomplete = int32(len(peermap.peers)) - complete

	return
}

// PeerList returns a peer list for the given hash capped at num
func (db *Memory) PeerList(h *shared.Hash, num int, noPeerID bool) []string {
	var i int

	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		return nil
	}

	peermap.RLock()
	maplen := len(peermap.peers)
	if num > maplen {
		num = maplen
	}

	peerList := make([]string, num)
	for id, peer := range peermap.peers {
		if i == num {
			break
		}
		dict := bencoding.NewDict()
		if noPeerID == false {
			dict.Any("peer id", string(id[:]))
		}
		dict.Any("ip", net.IP(peer.IP[:]).String())
		dict.Any("port", peer.Port)

		peerList[i] = dict.Get()
		i++
	}
	peermap.RUnlock()

	return peerList
}

// PeerListBytes returns a byte encoded peer list for the given hash capped at num
func (db *Memory) PeerListBytes(h *shared.Hash, num int) []byte {
	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		return nil
	}

	peermap.RLock()
	maplen := len(peermap.peers)
	if num > maplen {
		num = maplen
	}
	peerlist := make([]byte, 6*num)
	var pos int

	for _, peer := range peermap.peers {
		if pos/6 == num {
			break
		}

		copy(peerlist[pos:pos+4], peer.IP[:])
		binary.BigEndian.PutUint16(peerlist[pos+4:pos+6], peer.Port)
		pos += 6
	}
	peermap.RUnlock()

	return peerlist
}
