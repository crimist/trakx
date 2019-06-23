package tracker

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
	var peerList []string
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

// PeerListCompact returns the peer list byte encoded
func (h *Hash) PeerListCompact(num int64) string {
	var peerList string
	peerMap, _ := PeerDB[*h]

	var i int64
	for _, peer := range peerMap {
		if i == num {
			break
		}
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		// Network order
		binary.Write(writer, binary.BigEndian, utils.IPToInt(net.ParseIP(peer.IP)))
		binary.Write(writer, binary.BigEndian, peer.Port)
		writer.Flush()

		peerList += b.String()
		i++
	}

	return peerList
}

func (h *Hash) PeerListUDP(num int64) (peers []UDPPeer) {
	peerMap, _ := PeerDB[*h]

	var i int64
	for _, peer := range peerMap {
		if i == num {
			break
		}

		// TODO THIS NEEDS TO BE BIG ENDIAN
		p := UDPPeer{
			IP:   int32(utils.IPToInt(net.ParseIP(peer.IP))),
			Port: peer.Port,
		}

		peers = append(peers, p)
		i++
	}

	return
}
