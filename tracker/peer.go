package tracker

import (
	"bytes"
	"database/sql"
	"errors"

	"go.uber.org/zap"
)

// Peer holds peer information stores in the database
type Peer struct {
	ID       []byte `db:"id"`
	Key      []byte `db:"key"`
	Hash     Hash   `db:"hash"`
	IP       string `db:"ip"`
	Port     uint16 `db:"port"`
	Complete bool   `db:"complete"`
	LastSeen int64  `db:"last_seen"`
}

// InitPeer creates the peer db if it doesn't exist
func initPeerTable() {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS peers (`id` varbinary(255), `key` varbinary(255), `hash` varbinary(255), `ip` varchar(255), `port` int unsigned, `complete` boolean, `last_seen` bigint, PRIMARY KEY (`id`))")
	if err != nil {
		panic(err)
	}
}

// checkKey checks if the key is set and if it is valid
func (p *Peer) checkKey() error {
	var keydb []byte
	row := db.QueryRowx("SELECT `key` FROM peers WHERE id = ?", p.ID)
	if err := row.Scan(&keydb); err != nil && err != sql.ErrNoRows {
		logger.Error("checkKey err", zap.Error(err))
	}

	if bytes.Equal(p.Key, keydb) == false {
		logger.Info("invalid key",
			zap.String("ip", p.IP),
			zap.ByteString("db key", keydb),
			zap.ByteString("post key", p.Key),
		)
		return errors.New("Invalid key")
	}
	return nil
}

// Save creates or updates peer
func (p *Peer) Save() error {
	logger.Info("Save",
		zap.Any("Peer", p),
	)

	// Doesn't check if key matches atm
	insertpeer := "INSERT INTO peers (`id`, `key`, `hash`, `ip`, `port`, `complete`, `last_seen`) VALUES (?, ?, ?, ?, ?, ?, ?)" +
		"ON DUPLICATE KEY UPDATE id=VALUES(id), `key`=VALUES(`key`), hash=VALUES(hash), ip=VALUES(ip), port=VALUES(port), complete=VALUES(complete), last_seen=VALUES(last_seen) "

	if _, err := db.Exec(insertpeer, p.ID, p.Key, p.Hash, p.IP, p.Port, p.Complete, p.LastSeen); err != nil {
		return err
	}

	return nil
}

// Delete deletes peer
func (p *Peer) Delete() error {
	logger.Info("Delete",
		zap.Any("Peer", p),
	)
	if err := p.checkKey(); err != nil {
		return err
	}

	_, err := db.Exec("DELETE FROM peers WHERE id=?", p.ID)
	return err
}
