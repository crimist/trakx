package tracker

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/binary"
	"net"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/utils"
)

// Hash is the infohash of a torrent
type Hash []byte

// Banned checks if the hash is banned
func (h *Hash) Banned() bool {
	// var res int
	row := db.QueryRow("SELECT 1 FROM bans WHERE hash=?", h)
	if err := row.Scan(nil); err == sql.ErrNoRows {
		return false
	}

	return true
}

// Complete returns the number of peers that are complete
func (h *Hash) Complete() (int, error) {
	var count int

	row := db.QueryRow("SELECT count(*) FROM peers WHERE complete = true AND hash = ?", h)
	if err := row.Scan(&count); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return count, nil
}

// Incomplete returns the number of peers that are incomplete
func (h *Hash) Incomplete() (int, error) {
	var count int

	row := db.QueryRow("SELECT count(*) FROM peers WHERE complete = false AND hash = ?", h)
	if err := row.Scan(&count); err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return count, nil
}

// PeerList returns the peerlist bencoded
func (h *Hash) PeerList(num int64, noPeerID bool) ([]string, error) {
	var peerList []string
	var peers []Peer

	if err := db.Select(&peers, "SELECT id, ip, port FROM peers WHERE hash = ? LIMIT ?", h, num); err != nil {
		return peerList, err
	}
	for _, peer := range peers {
		dict := bencoding.NewDict()
		if noPeerID == false {
			dict.Add("peer id", peer.ID)
		}
		dict.Add("ip", peer.IP)
		dict.Add("port", peer.Port)

		peerList = append(peerList, dict.Get())
	}

	return peerList, nil
}

// PeerListCompact returns the peer list as byte encoded
func (h *Hash) PeerListCompact(num int64) (string, error) {
	var peerList string
	var peers []Peer

	if err := db.Select(&peers, "SELECT ip, port FROM peers WHERE hash = ? LIMIT ?", h, num); err != nil {
		return peerList, err
	}
	for _, peer := range peers {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		// Network order
		binary.Write(writer, binary.BigEndian, utils.IPToInt(net.ParseIP(peer.IP)))
		binary.Write(writer, binary.BigEndian, peer.Port)
		writer.Flush()

		peerList += b.String()
	}

	return peerList, nil
}
