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
	return len(db.db)
}

// HashStats returns number of complete and incomplete peers associated with the hash
func (db *PeerDatabase) HashStats(h *Hash) (complete, incomplete int32) {
	db.mu.RLock()
	peerMap, _ := db.db[*h]
	for _, peer := range peerMap {
		if peer.Complete == true {
			complete++
		} else {
			incomplete++
		}
	}
	db.mu.RUnlock()

	return
}

// PeerList returns a peer list for the given hash capped at num
func (db *PeerDatabase) PeerList(h *Hash, num int, noPeerID bool) []string {
	db.mu.RLock()
	peerMap, _ := db.db[*h]
	if num > len(peerMap) {
		num = len(peerMap)
	}
	peerList := make([]string, num)

	i := 0
	for id, peer := range peerMap {
		if i == num {
			break
		}
		dict := bencoding.NewDict()
		if noPeerID == false {
			dict.Add("peer id", string(id[:]))
		}
		dict.Add("ip", net.IP(peer.IP[:]).String())
		dict.Add("port", peer.Port)

		peerList[i] = dict.Get()
		i++
	}
	db.mu.RUnlock()

	return peerList
}

// PeerListBytes returns a byte encoded peer list for the given hash capped at num
func (db *PeerDatabase) PeerListBytes(h *Hash, num int) []byte {
	var peerList bytes.Buffer

	db.mu.RLock()
	peerMap, _ := db.db[*h]
	if num > len(peerMap) {
		num = len(peerMap)
	}
	peerList.Grow(6 * num)
	writer := bufio.NewWriterSize(&peerList, 6*num)

	for _, peer := range peerMap {
		if num == 0 {
			break
		}
		binary.Write(writer, binary.BigEndian, peer.IP)
		binary.Write(writer, binary.BigEndian, peer.Port)
		num--
	}
	db.mu.RUnlock()

	writer.Flush()
	return peerList.Bytes()
}
