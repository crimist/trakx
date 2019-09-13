package gomap

import (
	"database/sql"
	"errors"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/syc0x00/trakx/tracker/storage"
	"go.uber.org/zap"
)

const (
	maxrows = "10000" // -1 for unlimited
)

type PgBackup struct {
	pg *sql.DB
	db *Memory
}

func (bck *PgBackup) Init(db storage.Database) error {
	var err error

	bck.db = db.(*Memory)
	if bck.db == nil {
		panic("db nil on backup init")
	}

	bck.pg, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	err = bck.pg.Ping()
	if err != nil {
		bck.db.logger.Error("postgres ping() failed", zap.Error(err))
		return err
	}

	_, err = bck.pg.Exec("CREATE TABLE IF NOT EXISTS trakx (ts TIMESTAMP DEFAULT now(), bytes BYTEA)")
	if err != nil {
		bck.db.logger.Error("postgres table create failed", zap.Error(err))
		return err
	}

	return nil
}

func (bck PgBackup) save() error {
	data, err := bck.db.encode()
	if err != nil {
		return err
	}

	if data == nil {
		bck.db.logger.Error("Failed to encode db")
		return errors.New("Failed to encode db")
	}

	_, err = bck.pg.Query("INSERT INTO trakx(bytes) VALUES($1)", data)
	if err != nil {
		bck.db.logger.Error("postgres insert failed", zap.Error(err))
		return errors.New("postgres insert failed")
	}

	rm, err := bck.trimBackups()
	if err != nil {
		bck.db.logger.Error("failed to trim backups", zap.Error(err))
	} else {
		bck.db.logger.Info("Deleted expired postgres records", zap.Int64("deleted", rm))
	}

	bck.db.logger.Info("Deleted expired postgres records", zap.Int64("deleted", rm))

	return nil
}

func (bck PgBackup) SaveTmp() error {
	return bck.save()
}

func (bck PgBackup) SaveFull() error {
	return bck.save()
}

func (bck PgBackup) load() error {
	var data []byte

	err := bck.pg.QueryRow("SELECT bytes FROM trakx ORDER BY ts DESC LIMIT 1").Scan(&data)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			bck.db.logger.Info("Empty table")
			return nil
		}
		bck.db.logger.Error("SELECT failed", zap.Error(err))
		return nil
	}

	bck.db.decode(data)

	return nil
}

func (bck PgBackup) Load() error {
	err := bck.load()
	if err != nil {
		bck.db.logger.Error("failed to load pg db", zap.Error(err))
	}
	return err
}

func (bck PgBackup) trimBackups() (int64, error) {
	// delete records older than 7 days
	result, err := bck.pg.Exec("DELETE FROM trakx WHERE ts < NOW() - INTERVAL '7 days'")
	if err != nil {
		return -1, err
	}

	trimmed, err := result.RowsAffected()
	if err != nil {
		return -1, err
	}

	if maxrows != "-1" {
		result, err = bck.pg.Exec("DELETE FROM trakx WHERE ctid IN (SELECT ctid FROM trakx ORDER BY ctid DESC OFFSET " + maxrows + ")")
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
