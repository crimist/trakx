package tracker

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
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
	// open mysql conn
	db, err = sql.Open("mysql", "root@/bittorrent")
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// init logger
	if prod {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	defer logger.Sync()

	// init the banned hashes table
	initBanTable()

	return db, nil
}

// Clean auto cleans clients that haven't checked in in a certain amount of time
func Clean() {
	for {
		// Get all hash_ tables
		rows, err := db.Query("SELECT DISTINCT TABLE_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='bittorrent' AND TABLE_NAME LIKE 'Hash_%'")
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
			tables = append(tables, table)
		}
		rows.Close()

		// Clean them all
		for _, table := range tables {
			timeOut := int64(60 * 10) // 10 min
			query := fmt.Sprintf("DELETE FROM %s WHERE lastSeen < ?", table)
			_, err := db.Exec(query, time.Now().Unix()-timeOut)
			if err != nil {
				logger.Error(err.Error())
			}
		}

		logger.Info("Cleaned tables",
			zap.Int("table #", len(tables)),
			zap.Strings("tables", tables),
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
func NewTorrent(hash string) (*Torrent, BanErr) {
	t := Torrent{
		hash: EncodeInfoHash(hash),
	}

	// If it's a banned hash say fuck off
	isBanned := IsBannedHash(t.hash)
	if isBanned == -1 {
		return nil, Err
	} else if isBanned == 1 {
		// banned hash
		return nil, ErrBanned
	}

	if t.table() { // if it failed
		return nil, Err
	}
	return &t, ErrOK
}

func (t *Torrent) table() bool {
	_, err := db.Exec(fmt.Sprintf("SELECT 1 FROM %s LIMIT 1", t.hash))
	if err != nil { // If error the table doesn't exist
		// SQli safe because we change it to hex
		_, err = db.Exec(fmt.Sprintf("CREATE TABLE %s (id varchar(40), peerKey varchar(20), ip varchar(45), port smallint unsigned, complete bool, lastSeen bigint unsigned)", t.hash))
		if err != nil {
			logger.Error(err.Error())
			return true
		}
		logger.Info("New torrent",
			zap.String("hash", t.hash),
		)
	}
	return false
}

// Peer adds or updates a peer
func (t *Torrent) Peer(id string, key string, ip string, port string, complete bool) bool {
	query := fmt.Sprintf("UPDATE %s SET ip = ?, port = ?, complete = ?, lastSeen = ? WHERE id = ? AND peerKey = ?", t.hash)
	result, err := db.Exec(query, ip, port, complete, time.Now().Unix(), id, key)
	if err != nil {
		logger.Error(err.Error())
		return true
	}
	affected, err := result.RowsAffected()
	if err != nil {
		logger.Error(err.Error())
		return true
	}
	if affected == 0 { // They don't exist
		query := fmt.Sprintf("INSERT INTO %s VALUES (?, ?, ?, ?, ?, ?)", t.hash)
		_, err = db.Exec(query, id, key, ip, port, complete, time.Now().Unix())
		if err != nil {
			logger.Error(err.Error())
			return true
		}
		logger.Info("New peer",
			zap.String("id", id),
			zap.String("key", key),
			zap.String("ip", ip),
			zap.String("port", port),
			zap.Bool("complete", complete),
		)
	}
	return false
}

// RemovePeer removes the peer from the db
func (t *Torrent) RemovePeer(id string, key string) bool {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ? AND peerKey = ?", t.hash)
	_, err := db.Exec(query, id, key)
	if err != nil {
		logger.Error(err.Error())
		return true
	}
	return false
}

// GetPeerList gets numwant peers from the db
// If numwanst is unspecfied then it is all peers
func (t *Torrent) GetPeerList(num string) ([]string, bool) {
	if num == "" || num == "0" {
		num = "9999999" // Unlimited
	}

	query := fmt.Sprintf("SELECT id, ip, port FROM %s ORDER BY RAND() LIMIT ?", t.hash)
	rows, err := db.Query(query, num)
	if err != nil {
		logger.Error(err.Error())
		return nil, true
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
			return nil, true
		}
		peer := bencoding.NewDict()
		peer.Add("peer id", id)
		peer.Add("ip", ip)
		peer.Add("port", port)
		peerList = append(peerList, peer.Get())
	}
	return peerList, false
}

// GetPeerListCompact gets numwant peers from the db in the compact format
// If numwanst is unspecfied then it is all peers
func (t *Torrent) GetPeerListCompact(num string) (string, bool) {
	if num == "" || num == "0" {
		num = "9999999" // Unlimited
	}

	query := fmt.Sprintf("SELECT ip, port FROM %s ORDER BY RAND() LIMIT ?", t.hash)
	rows, err := db.Query(query, num)
	if err != nil {
		logger.Error(err.Error())
		return "", true
	}
	defer rows.Close()

	peerList := ""
	for rows.Next() {
		var ip string
		var port uint16
		err = rows.Scan(&ip, &port)
		if err != nil {
			logger.Error(err.Error())
			return "", true
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
	return peerList, false
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

// Error throws a tracker bencoded error
func Error(w http.ResponseWriter, reason string, status int) {
	d := bencoding.NewDict()
	d.Add("failure reason", reason)
	w.WriteHeader(status)
	fmt.Fprint(w, d.Get())
}

// InternalError is a wrapper to tell the client I fucked up
func InternalError(w http.ResponseWriter) {
	Error(w, "Internal Server Error", http.StatusInternalServerError)
}
