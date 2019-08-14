package shared

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"

	"github.com/syc0x00/trakx/bencoding"
)

// Hash is the infohash of a torrent
type Hash [20]byte

// Hashes gets the number of hashes
func (db *PeerDatabase) Hashes() int {
	return len(db.hashmap)
}

// HashStats returns number of complete and incomplete peers associated with the hash
func (db *PeerDatabase) HashStats(h *Hash) (complete, incomplete int32) {
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
func (db *PeerDatabase) PeerList(h *Hash, num int, noPeerID bool) []string {
	var i int

	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		return []string{}
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
func (db *PeerDatabase) PeerListBytes(h *Hash, num int) []byte {
	var peerList bytes.Buffer

	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		peerList.Bytes()
	}

	peermap.RLock()
	maplen := len(peermap.peers)
	if num > maplen {
		num = maplen
	}

	size := 6 * num
	peerList.Grow(size)
	writer := bufio.NewWriterSize(&peerList, size)

	for _, peer := range peermap.peers {
		if num == 0 {
			break
		}
		binary.Write(writer, binary.BigEndian, peer.IP)
		binary.Write(writer, binary.BigEndian, peer.Port)
		num--
	}
	peermap.RUnlock()

	writer.Flush()
	return peerList.Bytes()
}
