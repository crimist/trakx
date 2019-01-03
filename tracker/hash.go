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
type Hash []byte

// Banned checks if the hash is banned
func (h *Hash) Banned() bool {
	ban := Ban{}
	db.Where("hash = ?", h).First(&ban)

	// If Ban.Hash isn't nill than it's banned
	return (ban.Hash != nil)
}

// Complete returns the number of peers that are complete
func (h *Hash) Complete() (int, error) {
	var peers []Peer
	err := db.Where("complete = true AND hash = ?", h).Find(&peers).Error
	return len(peers), err
}

// Incomplete returns the number of peers that are incomplete
func (h *Hash) Incomplete() (int, error) {
	var peers []Peer
	err := db.Where("complete = false AND hash = ?", h).Find(&peers).Error
	return len(peers), err
}

// PeerList returns the peerlist bencoded
func (h *Hash) PeerList(num int64, noPeerID bool) ([]string, error) {
	var peerList []string
	var peers []Peer

	db.Where("hash = ?", h).Limit(num).Find(&peers)
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

// PeerListCompact returns the peer list as byte encoded
func (h *Hash) PeerListCompact(num int64) (string, error) {
	var peerList string
	var peers []Peer

	db.Where("hash = ?", h).Limit(num).Find(&peers)
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
