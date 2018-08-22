package tracker

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver

	"github.com/Syc0x00/Trakx/bencoding"
)

var db *sql.DB

// Init x
func Init() (*sql.DB, error) {
	var err error
	db, err = sql.Open("mysql", "root@/bittorrent")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Clean x
// Have a lastseen timestamp in the mysql DB and if we havn't seen them in 1 hour
// Remove them
func Clean() {
	for {
		// Cleanup
		time.Sleep(5 * time.Minute)
	}
}

// Torrent x
type Torrent struct {
	hash string
}

// NewTorrent x
func NewTorrent(hash string) (Torrent, error) {
	t := Torrent{
		hash: fmt.Sprintf("%X", hash),
	}
	err := t.table()
	return t, err
}

func (t *Torrent) table() error {
	_, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", t.hash))
	if err != nil { // If error the table doesn't exist
		// SQli safe because we change it to hex
		_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s (id varchar(40), peerKey varchar(20), ip varchar(255), port smallint unsigned, complete bool)", t.hash))
		return err
	}
	return nil
}

// Peer adds or updates a peer
func (t *Torrent) Peer(id string, key string, ip string, port string, complete bool) error {
	query := fmt.Sprintf("UPDATE %s SET ip = ?, port = ?, complete = ? WHERE id = ? AND peerKey = ?)", t.hash)
	_, err := db.Exec(query, ip, port, complete, id, key)
	if err != nil {
		// They don't exist
		query := fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?, ?, ?)", t.hash)
		_, err = db.Exec(query, id, key, ip, port, complete)
	}
	return err
}

// RemovePeer removes the peer from the db
func (t *Torrent) RemovePeer(id string, key string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ? AND peerKey = ?", t.hash)
	_, err := db.Exec(query, id, key)
	return err
}

// GetPeerList gets numwant peers from the db
// If numwanst is unspecfied then it is all peers
func (t *Torrent) GetPeerList(num string) ([]string, error) {
	if num == "" || num == "0" {
		num = "9999999" // Unlimited
	}

	query := fmt.Sprintf("SELECT id, ip, port FROM %s ORDER BY RAND() LIMIT ?", t.hash)
	rows, err := db.Query(query, num)
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

// GetPeerListCompact gets numwant peers from the db in the compact format
// If numwanst is unspecfied then it is all peers
func (t *Torrent) GetPeerListCompact(num string) (string, error) {
	if num == "" || num == "0" {
		num = "9999999" // Unlimited
	}

	query := fmt.Sprintf("SELECT id, ip, port FROM %s ORDER BY RAND() LIMIT ?", t.hash)
	rows, err := db.Query(query, num)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	peerList := ""
	for rows.Next() {
		var id string
		var ip string
		var port uint16
		err = rows.Scan(&id, &ip, &port)
		if err != nil {
			return "", err
		}

		// Network order
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)
		binary.Write(writer, binary.BigEndian, net.ParseIP(ip))
		binary.Write(writer, binary.BigEndian, port)
		peerList += b.String()
	}
	return peerList, nil
}

// Complete returns the number of peers that are complete
func (t *Torrent) Complete() (int, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE complete = true", t.hash))
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
	rows, err := db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE complete = false", t.hash))
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

// Error throws a tracker bencoded error and writes to stdout with
// a stack trace.
func Error(w io.Writer, reason string) {
	fmt.Println("Err:", reason)
	debug.PrintStack()

	d := bencoding.NewDict()
	d.Add("failure reason", reason)
	fmt.Fprint(w, d.Get())
}

// Delete soon if 100% unneeded

/*
// NewPeer creates a new peer in the dp
func (t *Torrent) NewPeer(id string, key string, ip string, port string, complete bool) error {
	query := fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?, ?, ?)", t.hash)
	_, err := db.Exec(query, id, key, ip, port, complete)
	if err != nil {
		return err
	}
	return nil
}

// UpdatePeer updates an existing peer in the db
func (t *Torrent) UpdatePeer(id string, key string, ip string, port string, complete bool) error {
	query := fmt.Sprintf("UPDATE %s SET id = ?, ip = ?, port = ?, complete = ? WHERE peerKey = ?)", t.hash)
	_, err := db.Exec(query, id, ip, port, complete, key)
	if err != nil {
		return err
	}
	return nil
}
*/
