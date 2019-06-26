package shared

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/utils"
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
func (h *Hash) PeerList(num int64, noPeerID bool) []string {
	peerList := make([]string, num)
	peerMap, _ := PeerDB[*h]

	var i int64
	for id, peer := range peerMap {
		if i == num {
			break
		}
		dict := bencoding.NewDict()
		if noPeerID == false {
			dict.Add("peer id", id)
		}
		dict.Add("ip", peer.IP)
		dict.Add("port", peer.Port)

		peerList = append(peerList, dict.Get())
		i++
	}

	return peerList
}

// PeerListBytes returns the peer list byte encoded
func (h *Hash) PeerListBytes(num int64) []byte {
	var peerList bytes.Buffer
	peerList.Grow(6 * int(num))
	writer := bufio.NewWriter(&peerList)
	peerMap, _ := PeerDB[*h]

	for _, peer := range peerMap {
		if num == 0 {
			break
		}

		binary.Write(writer, binary.BigEndian, utils.IPToInt(net.ParseIP(peer.IP)))
		binary.Write(writer, binary.BigEndian, peer.Port)
		num--
	}

	writer.Flush()
	return peerList.Bytes()
}
