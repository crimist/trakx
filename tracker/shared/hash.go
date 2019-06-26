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

// PeerListCompact returns the peer list byte encoded
func (h *Hash) PeerListCompact(num int64) string {
	var peerList bytes.Buffer
	peerList.Grow(6 * int(num))
	writer := bufio.NewWriter(&peerList)
	peerMap, _ := PeerDB[*h]

	var i int64
	for _, peer := range peerMap {
		if i == num {
			break
		}

		// Network order
		binary.Write(writer, binary.BigEndian, utils.IPToInt(net.ParseIP(peer.IP)))
		binary.Write(writer, binary.BigEndian, peer.Port)
		i++
	}

	writer.Flush()
	return peerList.String()
}

func (h *Hash) PeerListUDP(num int32) (peers []UDPPeer) {
	peerMap, _ := PeerDB[*h]

	for _, peer := range peerMap {
		if num == 0 {
			break
		}

		p := UDPPeer{
			IP:   int32(utils.IPToInt(net.ParseIP(peer.IP))),
			Port: peer.Port,
		}

		// TODO: Find less ugly way to convert to big endian
		p.Port = (p.Port << 8) | (p.Port >> 8)
		p.IP = (p.IP << 24) | ((p.IP << 8) & 0x00ff0000) | ((p.IP >> 8) & 0x0000ff00) | (p.IP >> 24)

		peers = append(peers, p)
		num--
	}

	return
}
