package inmemory

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"os"

	_ "github.com/lib/pq"
	"github.com/syc0x00/trakx/tracker/database"
	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type PgBackup struct {
	pg *sql.DB
	db *Memory
}

func (bck *PgBackup) Init(db database.Database) error {
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

	_, err = bck.pg.Exec("CREATE TABLE IF NOT EXISTS peerdb (ts TIMESTAMP DEFAULT now(), bytes TEXT)")
	if err != nil {
		bck.db.logger.Error("postgres table create failed", zap.Error(err))
		return err
	}

	return nil
}

func (bck PgBackup) save() error {
	data := bck.db.encode()
	if data == nil {
		bck.db.logger.Error("Failed to encode db")
		return errors.New("Failed to encode db")
	}

	_, err := bck.pg.Query("INSERT INTO peerdb(bytes) VALUES($1)", data)
	if err != nil {
		bck.db.logger.Error("postgres insert failed", zap.Error(err))
		return errors.New("postgres insert failed")
	}

	rm, err := bck.trimBackups()
	if err != nil {
		bck.db.logger.Error("failed to trim backups", zap.Error(err))
		return errors.New("failed to trim backups")
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
	err := bck.pg.QueryRow("SELECT bytes FROM peerdb ORDER BY ts DESC LIMIT 1").Scan(&data)
	if err != nil {
		return err
	}

	var hash shared.Hash
	bck.db.make()

	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	for _, file := range archive.File {
		hashbytes, err := hex.DecodeString(file.Name)
		if err != nil {
			return err
		}
		copy(hash[:], hashbytes)
		peermap := bck.db.makePeermap(&hash)

		reader, err := file.Open()
		if err != nil {
			return err
		}
		err = gob.NewDecoder(reader).Decode(&peermap.peers)
		if err != nil {
			return err
		}
		reader.Close()
	}

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
	result, err := bck.pg.Exec("DELETE FROM peerdb WHERE ts < NOW() - INTERVAL '7 days'")
	if err != nil {
		return -1, err
	}
	return result.RowsAffected()
}
