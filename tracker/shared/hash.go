package shared

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"

	"github.com/Syc0x00/Trakx/bencoding"
)

// Hash is the infohash of a torrent
type Hash [20]byte

// Complete returns number of complete and incomplete peers associated with the hash
func (h *Hash) Complete() (complete, incomplete int32) {
	peerMap, _ := PeerDB[*h]

	for _, peer := range peerMap {
		if peer.Complete == true {
			complete++
		} else {
			incomplete++
		}
	}

	return complete, incomplete
}

// PeerList returns the peerlist bencoded
func (h *Hash) PeerList(num int, noPeerID bool) []string {
	peerMap, _ := PeerDB[*h]
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

	return peerList
}

// PeerListBytes returns the peer list byte encoded
func (h *Hash) PeerListBytes(num int) []byte {
	peerMap, _ := PeerDB[*h]
	if num > len(peerMap) {
		num = len(peerMap)
	}
	var peerList bytes.Buffer
	writer := bufio.NewWriter(&peerList)
	peerList.Grow(6 * int(num))

	for _, peer := range peerMap {
		if num == 0 {
			break
		}

		binary.Write(writer, binary.BigEndian, peer.IP)
		binary.Write(writer, binary.BigEndian, peer.Port)
		num--
	}

	writer.Flush()
	return peerList.Bytes()
}
