package database

import (
	"os"
	"time"

	"github.com/crimist/trakx/storage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const defaultFilePermission = 0640 // rw-r-----

type FilePersistance struct{}

func (fp *FilePersistance) write(db *Database, path string) error {
	zap.L().Info("Persisting database to file", zap.String("path", path))
	start := time.Now()

	encoded, err := encodeBinary(db)
	if err != nil {
		return errors.Wrap(err, "failed to binary encode database")
	}
	os.WriteFile(path, encoded, defaultFilePermission)

	zap.L().Info("Persisted database to file", zap.Duration("elapsed", time.Since(start)))
	return nil
}

func (fp *FilePersistance) read(db *Database, path string) error {
	zap.L().Info("Loading database from file", zap.String("path", path))
	start := time.Now()

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to read file from disk")
	}

	db.torrents = make(map[storage.Hash]*Torrent, 1) // TODO: run some calculations to estimate the size of the map
	peers, torrents, err := decodeBinary(db, data)
	if err != nil {
		return errors.Wrap(err, "failed to binary decode database")
	}

	zap.L().Info("Loaded database from file", zap.Int("peers", peers), zap.Int("hashes", torrents), zap.Duration("elapsed", time.Since(start)))
	return nil
}
