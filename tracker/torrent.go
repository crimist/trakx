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
func Complete(hash []byte) (int, error) {
	var peers []Peer
	err := db.Where("complete = true AND hash = ?", hash).Find(&peers).Error
	return len(peers), err
}

// Incomplete returns the number of peers that are incomplete
func Incomplete(hash []byte) (int, error) {
	var peers []Peer
	err := db.Where("complete = false AND hash = ?", hash).Find(&peers).Error
	return len(peers), err
}

// PeerList x
func PeerList(hash []byte, num int64, noPeerID bool) ([]string, error) {
	var peerList []string
	var peers []Peer

	db.Where("hash = ?", hash).Limit(num).Find(&peers)
	for _, peer := range peers {
		dict := bencoding.NewDict()
		// if they don't want peerid
		if noPeerID == false {
			dict.Add("peer id", peer.ID)
		}
		dict.Add("ip", peer.IP)
		dict.Add("port", peer.Port)
		peerList = append(peerList, dict.Get())
	}

	return peerList, db.Error
}

// PeerListCompact x
func PeerListCompact(hash []byte, num int64) (string, error) {
	var peerList string
	var peers []Peer

	db.Where("hash = ?", hash).Limit(num).Find(&peers)
	for _, peer := range peers {
		// Network order
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		binary.Write(writer, binary.BigEndian, utils.IPToInt(net.ParseIP(peer.IP)))
		binary.Write(writer, binary.BigEndian, peer.Port)
		writer.Flush()

		peerList += b.String()
	}

	return peerList, db.Error
}
