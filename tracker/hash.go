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
	return false
}

func (h *Hash) Complete() (int, int) {
	complete := 0
	incomplete := 0


	for _, val := range db {
		if bytes.Equal(val.Hash, *h) {
			if val.Complete == true {
				complete++
			} else {
				incomplete++
			}
		}
	}

	return complete, incomplete
}

// PeerList returns the peerlist bencoded
func (h *Hash) PeerList(num int64, noPeerID bool) ([]string) {
	var peerList []string

	for id, peer := range db {
		if bytes.Equal(peer.Hash, *h) {
			dict := bencoding.NewDict()
			if noPeerID == false {
				dict.Add("peer id", id)
			}
			dict.Add("ip", peer.IP)
			dict.Add("port", peer.Port)

			peerList = append(peerList, dict.Get())
		}
	}

	return peerList
}

// PeerListCompact returns the peer list as byte encoded
func (h *Hash) PeerListCompact(num int64) (string) {
	var peerList string


	for _, peer := range db {
		if bytes.Equal(peer.Hash, *h) {
			var b bytes.Buffer
			writer := bufio.NewWriter(&b)

			// Network order
			binary.Write(writer, binary.BigEndian, utils.IPToInt(net.ParseIP(peer.IP)))
			binary.Write(writer, binary.BigEndian, peer.Port)
			writer.Flush()

			peerList += b.String()
		}
	}

	return peerList
}
