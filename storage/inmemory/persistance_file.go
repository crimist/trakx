package inmemory

import (
	"os"
	"time"

	"github.com/crimist/trakx/storage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const filePermission = 0640 // rw-r-----

type FilePersistance struct{}

func (fp *FilePersistance) write(db *InMemory, filepath string) error {
	zap.L().Info("persisting database to file")
	start := time.Now()

	encoded, err := encodeBinary(db)
	if err != nil {
		return errors.Wrap(err, "failed to binary encode databse")
	}
	os.WriteFile(filepath, encoded, filePermission)

	zap.L().Info("persisted database to file", zap.Duration("elapsed", time.Since(start)))
	return nil
}

func (fp *FilePersistance) read(db *InMemory, filepath string) error {
	zap.L().Info("loading databse from file")
	start := time.Now()

	data, err := os.ReadFile(filepath)
	if err != nil {
		return errors.Wrap(err, "failed to read file from disk")
	}

	db.torrents = make(map[storage.Hash]*Torrent, 1) // TODO: do some math to estimate the size of the map
	peers, torrents, err := decodeBinary(db, data)
	if err != nil {
		return errors.Wrap(err, "failed to binary decode database")
	}

	zap.L().Info("loaded database from file", zap.Int("peers", peers), zap.Int("hashes", torrents), zap.Duration("elapsed", time.Since(start)))
	return nil
}
