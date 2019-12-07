package gomap

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	"github.com/crimist/trakx/tracker/storage"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	// Maximum retention for entries. Rows older than this will be removed
	// "off" to disable
	maxDate = "7 days"

	// Maximum number of rows. Rows exceeding this will be removed by timestamp
	// -1 for unlimited
	maxRows = "10"
)

type PgBackup struct {
	pg *sql.DB
	db *Memory
}

func (bck *PgBackup) Init(db storage.Database) error {
	var err error

	bck.db = db.(*Memory)
	if bck.db == nil {
		return errors.New("nil database on backup Init()")
	}

	bck.pg, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return errors.New("Failed to open pg connection: " + err.Error())
	}

	err = bck.pg.Ping()
	if err != nil {
		return errors.New("postgres ping() failed: " + err.Error())
	}

	_, err = bck.pg.Exec("CREATE TABLE IF NOT EXISTS trakx (ts TIMESTAMP DEFAULT now(), bytes BYTEA)")
	if err != nil {
		return errors.New("Failed to CREATE TABLE: " + err.Error())
	}

	return nil
}

func (bck PgBackup) save() error {
	data, err := bck.db.encodeBinary()
	if err != nil {
		bck.db.conf.Logger.Error("Failed to encode", zap.Error(err))
		return err
	}

	_, err = bck.pg.Query("INSERT INTO trakx(bytes) VALUES($1)", data)
	if err != nil {
		bck.db.conf.Logger.Error("postgres insert failed", zap.Error(err))
		return errors.New("postgres insert failed")
	}

	rm, err := bck.trim()
	if err != nil {
		bck.db.conf.Logger.Error("failed to trim backups", zap.Error(err))
	} else {
		bck.db.conf.Logger.Info("Deleted expired postgres records", zap.Int64("deleted", rm))
	}

	bck.db.conf.Logger.Info("Deleted expired postgres records", zap.Int64("deleted", rm))

	return nil
}

func (bck PgBackup) Save() error {
	return bck.save()
}

func (bck PgBackup) load() error {
	var data []byte

	err := bck.pg.QueryRow("SELECT bytes FROM trakx ORDER BY ts DESC LIMIT 1").Scan(&data)
	if err != nil {
		defer bck.db.make()

		if strings.Contains(err.Error(), "no rows in result set") { // empty postgres table
			bck.db.conf.Logger.Info("No stored database found")
			return nil
		}
		return errors.New("postgres SELECT query failed: " + err.Error())
	}

	bck.db.conf.Logger.Info("Loading stored database", zap.Int("size", len(data)))
	peers, hashes, err := bck.db.decodeBinary(data)
	if err != nil {
		bck.db.conf.Logger.Error("Error decoding stored database", zap.Error(err))
		bck.db.make()
		return err
	}

	bck.db.conf.Logger.Info("Loaded stored database", zap.Int("peers", peers), zap.Int("hashes", hashes))
	return nil
}

func (bck PgBackup) Load() error {
	return bck.load()
}

func (bck PgBackup) trim() (int64, error) {
	var trimmed int64

	if maxDate != "off" {
		result, err := bck.pg.Exec("DELETE FROM trakx WHERE ts < NOW() - INTERVAL '" + maxDate + "'")
		if err != nil {
			return -1, err
		}

		trimmed, err = result.RowsAffected()
		if err != nil {
			return -1, err
		}
	}

	if maxRows != "-1" {
		result, err := bck.pg.Exec("DELETE FROM trakx WHERE ctid IN (SELECT ctid FROM trakx ORDER BY ctid DESC OFFSET " + maxRows + ")")
		if err != nil {
			return -1, err
		}
		removedRows, err := result.RowsAffected()
		if err != nil {
			return -1, err
		}

		trimmed += removedRows
	}

	return trimmed, nil
}
