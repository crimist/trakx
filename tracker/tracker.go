package tracker

import (
	"database/sql"
	"fmt"
	"io"
	"runtime/debug"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver

	"github.com/Syc0x00/Trakx/bencoding"
)

var db *sql.DB

// Init x
func Init() *sql.DB {
	var err error
	db, err = sql.Open("mysql", "root@/bittorrent")
	if err != nil {
		fmt.Println(err)
	}
	if err := db.Ping(); err != nil {
		fmt.Println(err)
	}
	return db
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
	fmt.Println(t.hash)
	_, err := db.Exec("SELECT 1 FROM ? LIMIT 1", t.hash)
	if err != nil { // If error table doesn't exist
		_, err = db.Exec("CREATE TABLE ? (id varchar(40), peerKey varchar(20), ip varchar(255), port smallint unsigned, complete bool)", t.hash)
		return err
	}
	return nil
}

// NewPeer creates a new peer in the dp
func (t *Torrent) NewPeer(id string, key string, ip string, port string, complete bool) error {
	_, err := db.Exec("INSERT INTO $1 VALUES ($2, $3, $4, $5, $6)", t.hash, id, key, ip, port, complete)
	if err != nil {
		return err
	}
	return nil
}

// UpdatePeer updates an existing peer in the db
func (t *Torrent) UpdatePeer(id string, key string, ip string, port string, complete bool) error {
	_, err := db.Exec("UPDATE $1 SET id = $2, ip = $3, port = $4, complete = $5 WHERE key = $6)", t.hash, id, ip, port, complete, key)
	if err != nil {
		return err
	}
	return nil
}

// RemovePeer removes the peer from the db
func (t *Torrent) RemovePeer(id string, key string) error {
	_, err := db.Exec("DELETE FROM $1 WHERE id = $2 AND key = $4", t.hash, id, key)
	if err != nil {
		return err
	}
	return nil
}

// GetPeerList gets numwant peers from the db
// If numwanst is unspecfied then it is all peers
func (t *Torrent) GetPeerList(num string) ([]string, error) {
	if num == "" || num == "0" {
		num = "9999999" // Unlimited
	}

	rows, err := db.Query("SELECT id, ip, port FROM $1 ORDER BY RAND() LIMIT $2", t.hash, num)
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

// Error throws a tracker bencoded error
func Error(w io.Writer, reason string) {
	fmt.Println("Err:", reason)
	debug.PrintStack()

	d := bencoding.NewDict()
	d.Add("failure reason", reason)
	fmt.Fprint(w, d.Get())
}
