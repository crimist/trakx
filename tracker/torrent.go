package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/utils"
	"go.uber.org/zap"
)

// Torrent x
type Torrent struct {
	hash string
}

// NewTorrent x
func NewTorrent(hash string) (Torrent, TrackErr) {
	t := Torrent{
		hash: EncodeInfoHash(hash),
	}

	// If it's a banned hash say fuck off
	tErr := IsBannedHash(t.hash)
	if tErr == Error {
		return t, Error
	} else if tErr == Banned {
		return t, Banned
	}

	if t.table() != OK {
		return t, Error
	}
	return t, OK
}

// table creates a table for the torrents hash if it doesn't exist
func (t *Torrent) table() TrackErr {
	_, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", t.hash))

	// If error the table doesn't exist
	if err != nil {
		_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s (id varchar(40), peerKey varchar(20), ip varchar(45), port smallint unsigned, complete bool, lastSeen bigint unsigned)", t.hash))
		if err != nil {
			logger.Error(err.Error())
			return Error
		}

		logger.Info("New torrent",
			zap.String("hash", t.hash),
		)
	}

	return OK
}

// Peer adds or updates a peer
func (t *Torrent) Peer(id string, key string, ip string, port uint16, complete bool) TrackErr {
	// Update peer
	query := fmt.Sprintf("UPDATE %s SET ip = ?, port = ?, complete = ?, lastSeen = ? WHERE id = ? AND peerKey = ?", t.hash)
	result, err := db.Exec(query, ip, port, complete, time.Now().Unix(), id, key)
	if err != nil {
		logger.Error(err.Error())
		return Error
	}

	// Check if anyone was updated
	affected, err := result.RowsAffected()
	if err != nil {
		logger.Error(err.Error())
		return Error
	}

	// If they don't exist create the peer
	if affected == 0 {
		query := fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?, ?, ?, ?)", t.hash)
		_, err = db.Exec(query, id, key, ip, port, complete, time.Now().Unix())
		if err != nil {
			logger.Error(err.Error())
			return Error
		}

		logger.Info("New peer",
			zap.String("id", id),
			zap.String("key", key),
			zap.String("ip", ip),
			zap.Uint16("port", port),
			zap.Bool("complete", complete),
		)
	}

	return OK
}

// RemovePeer removes the peer from the db
func (t *Torrent) RemovePeer(id string, key string) TrackErr {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ? AND peerKey = ?", t.hash)
	_, err := db.Exec(query, id, key)
	if err != nil {
		logger.Error(err.Error())
		return Error
	}

	return OK
}

// GetPeerList gets numwant peers from the db
func (t *Torrent) GetPeerList(num uint64) ([]string, TrackErr) {
	// If they don't specify how many peers they want default to all
	if num == 0 {
		num = 18446744073709551615
	}

	query := fmt.Sprintf("SELECT id, ip, port FROM %s ORDER BY RAND() LIMIT ?", t.hash)
	rows, err := db.Query(query, num)
	if err != nil {
		logger.Error(err.Error())
		return nil, Error
	}
	defer rows.Close()

	peerList := []string{}
	for rows.Next() {
		var id string
		var ip string
		var port uint16
		err = rows.Scan(&id, &ip, &port)
		if err != nil {
			logger.Error(err.Error())
			return nil, Error
		}
		peer := bencoding.NewDict()
		peer.Add("peer id", id)
		peer.Add("ip", ip)
		peer.Add("port", port)
		peerList = append(peerList, peer.Get())
	}

	return peerList, OK
}

// GetPeerListCompact gets numwant peers from the db in the compact format
func (t *Torrent) GetPeerListCompact(num uint64) (string, TrackErr) {
	// If they don't specify how many peers they want default to all
	if num == 0 {
		num = 18446744073709551615
	}

	query := fmt.Sprintf("SELECT ip, port FROM %s ORDER BY RAND() LIMIT ?", t.hash)
	rows, err := db.Query(query, num)
	if err != nil {
		logger.Error(err.Error())
		return "", Error
	}
	defer rows.Close()

	peerList := ""
	for rows.Next() {
		var ip string
		var port uint16
		err = rows.Scan(&ip, &port)
		if err != nil {
			logger.Error(err.Error())
			return "", Error
		}

		// Network order
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		ipBytes := utils.IPToInt(net.ParseIP(ip))

		binary.Write(writer, binary.BigEndian, ipBytes)
		binary.Write(writer, binary.BigEndian, port)
		writer.Flush()
		peerList += b.String()
	}

	return peerList, OK
}

// Complete returns the number of peers that are complete
// returns -1 on failure
func (t *Torrent) Complete() int {
	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE complete = true", t.hash))
	if err != nil {
		logger.Error(err.Error())
		return -1
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			logger.Error(err.Error())
			return -1
		}
	}

	return count
}

// Incomplete returns the number of peers that are incomplete
// returns -1 on failure
func (t *Torrent) Incomplete() int {
	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE complete = false", t.hash))
	if err != nil {
		logger.Error(err.Error())
		return -1
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			logger.Error(err.Error())
			return -1
		}
	}

	return count
}
