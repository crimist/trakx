package tracker

import (
	"database/sql"
	"fmt"

	"github.com/Syc0x00/Trakx/bencoding"
)

var db *sql.DB

// Init x
func Init() {
	db, err := sql.Open("mysql", "user:password@/")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Println(err)
	}

	_, err = db.Exec("CREATE DATABASE bittorrent")
	if err != nil {
		fmt.Println(err)
	}
	_, err = db.Exec("USE bittorrent")
	if err != nil {
		fmt.Println(err)
	}
}

// Torrent x
type Torrent struct {
	hash string
}

// NewTorrent x
func NewTorrent(hash string) (Torrent, error) {
	t := Torrent{
		hash: hash,
	}
	err := t.table()
	return t, err
}

func (t *Torrent) table() error {
	_, err := db.Exec("SELECT 1 FROM %s LIMIT 1", t.hash)
	if err != nil { // If error table doesn't exist
		_, err = db.Exec("CREATE TABLE ? (id varchar(20), key varchar(20), ip varchar(255), port unsigned smallint, bool complete)", t.hash)
		return err
	}
	return nil
}

// NewPeer creates a new peer in the dp
func (t *Torrent) NewPeer(id string, key string, ip string, port string, complete bool) error {
	_, err := db.Exec("INSERT INTO ? VALUES (?, ?, ?, ?, ?)", t.hash, id, key, ip, port, complete)
	if err != nil {
		return err
	}
	return nil
}

// UpdatePeer updates an existing peer in the db
func (t *Torrent) UpdatePeer(id string, key string, ip string, port string, complete bool) error {
	_, err := db.Exec("UPDATE ? SET id = ?, ip = ?, port = ?, complete = ? WHERE key = ?)", t.hash, id, ip, port, complete, key)
	if err != nil {
		return err
	}
	return nil
}

// RemovePeer removes the peer from the db
func (t *Torrent) RemovePeer(key string) error {
	_, err := db.Exec("DELETE FROM ? WHERE key = ?", t.hash, key)
	if err != nil {
		return err
	}
	return nil
}

// GetPeerList gets numwant peers from the db
// If numwanst is unspecfied then it is all peers
func (t *Torrent) GetPeerList(num string) ([]string, error) {
	if num == "" {
		num = "9999999" // Unlimited
	}

	rows, err := db.Query("SELECT id, ip, port FROM ? ORDER BY RAND() LIMIT ?", t.hash, num)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	peerList := []string{}
	for rows.Next() {
		var id string
		var ip string
		var port uint16
		err = rows.Scan(&id, &ip, &port)
		if err != nil {
			return nil, err
		}
		peer := bencoding.NewDict()
		peer.Add("peer id", id)
		peer.Add("ip", ip)
		peer.Add("port", port)
		peerList = append(peerList, peer.Get())
	}
	return peerList, nil
}

// Complete returns the number of peers that are complete
func (t *Torrent) Complete() (int, error) {
	rows, err := db.Query("SELECT COUNT(*) FROM ? WHERE complete = true", t.hash)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count, nil
}

// Incomplete returns the number of peers that are incomplete
func (t *Torrent) Incomplete() (int, error) {
	rows, err := db.Query("SELECT COUNT(*) FROM ? WHERE complete = false", t.hash)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, err
		}
	}
	return count, nil
}
