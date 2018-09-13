package tracker

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"go.uber.org/zap"

	"github.com/Syc0x00/Trakx/bencoding"
	"github.com/Syc0x00/Trakx/utils"
)

var (
	db     *sql.DB
	logger *zap.Logger
)

// Init x
func Init(prod bool) (*sql.DB, error) {
	var err error
	db, err = sql.Open("mysql", "root@/bittorrent")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	if prod {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	defer logger.Sync()

	return db, nil
}

// Clean auto cleans clients that haven't checked in in a certain amount of time
func Clean() {
	for {
		// Get sql table list
		rows, err := db.Query("SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='bittorrent'")
		if err != nil {
			logger.Error(err.Error())
		}
		tables := []string{}
		for rows.Next() {
			var table string
			err = rows.Scan(&table)
			if err != nil {
				logger.Error(err.Error())
			}
			tables = append(tables)
		}
		rows.Close()

		// Clean them all
		deleted := 0
		for _, table := range tables {
			timeOut := int64(60 * 10) // 10 min
			query := fmt.Sprintf("DELETE FROM %s WHERE lastSeen < ?", table)
			result, err := db.Exec(query, time.Now().Unix()-timeOut)
			if err != nil {
				logger.Error(err.Error())
			}
			affected, err := result.RowsAffected()
			if err != nil {
				logger.Error(err.Error())
			}
			deleted += int(affected)
		}

		logger.Info("Cleaned tables",
			zap.Int("affected", deleted),
		)

		// Auto delete empty tables
		_, err = db.Exec("SELECT CONCAT('DROP TABLE ', GROUP_CONCAT(table_name), ';') AS query FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_ROWS = '0' AND TABLE_SCHEMA = 'bittorrent'")
		if err != nil {
			logger.Error(err.Error())
		}

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
		_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s (id varchar(40), peerKey varchar(20), ip varchar(45), port smallint unsigned, complete bool, lastSeen bigint unsigned)", t.hash))
		return err
	}
	return nil
}

// Peer adds or updates a peer
func (t *Torrent) Peer(id string, key string, ip string, port string, complete bool) error {
	query := fmt.Sprintf("UPDATE %s SET ip = ?, port = ?, complete = ?, lastSeen = ? WHERE id = ? AND peerKey = ?", t.hash)
	result, err := db.Exec(query, ip, port, complete, time.Now().Unix(), id, key)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 { // They don't exist
		query := fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?, ?, ?, ?)", t.hash)
		_, err = db.Exec(query, id, key, ip, port, complete, time.Now().Unix())
		logger.Info("New peer",
			zap.String("id", id),
			zap.String("key", key),
			zap.String("ip", ip),
			zap.String("port", port),
			zap.Bool("complete", complete),
		)
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

	query := fmt.Sprintf("SELECT ip, port FROM %s ORDER BY RAND() LIMIT ?", t.hash)
	rows, err := db.Query(query, num)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	peerList := ""
	for rows.Next() {
		var ip string
		var port uint16
		err = rows.Scan(&ip, &port)
		if err != nil {
			return "", err
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
	logger.Error("Error: ",
		zap.String("reason", reason),
	)

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
