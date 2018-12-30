package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/utils"
)

// Complete returns the number of peers that are complete
func Complete() (int, error) {
	var peers []Peer
	db.Where("complete = true").Find(&peers)
	return len(peers), db.Error
}

// Incomplete returns the number of peers that are incomplete
func Incomplete() (int, error) {
	var peers []Peer
	db.Where("complete = false").Find(&peers)
	return len(peers), db.Error
}

// PeerList x
func PeerList(num int64) ([]string, error) {
	var peerList []string
	var peers []Peer
	if num == 0 {
		num = 200 // If not specified default to 200
	}

	db.Limit(num).Find(&peers)
	for _, peer := range peers {
		dict := bencoding.NewDict()
		dict.Add("peer id", peer.ID)
		dict.Add("ip", peer.IP)
		dict.Add("port", peer.Port)
		peerList = append(peerList, dict.Get())
	}

	return peerList, db.Error
}

// PeerListCompact x
func PeerListCompact(num int64) (string, error) {
	var peerList string
	var peers []Peer
	if num == 0 {
		num = 200 // If not specified default to 200
	}

	db.Limit(num).Find(&peers)
	for _, peer := range peers {
		// Network order
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)
		ipBytes := utils.IPToInt(net.ParseIP(peer.IP))

		binary.Write(writer, binary.BigEndian, ipBytes)
		binary.Write(writer, binary.BigEndian, peer.Port)
		writer.Flush()

		peerList += b.String()
	}

	return peerList, db.Error
}
