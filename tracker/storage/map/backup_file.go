package gomap

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type FileBackup struct {
	db *Memory
}

func (bck *FileBackup) Init(db storage.Database) error {
	bck.db = db.(*Memory)
	if bck.db == nil {
		panic("db nil on backup init")
	}
	return nil
}

func (bck *FileBackup) Load() error {
	config.Logger.Info("Loading database from file")
	start := time.Now()
	path := config.CacheDir + "peers.db"

	_, err := os.Stat(path)
	if err != nil {
		// If the file doesn't exist than create an empty database and return success
		if os.IsNotExist(err) {
			bck.db.make()
			config.Logger.Info("Database file not found, created empty database", zap.String("filepath", path))
			return nil
		}

		return errors.Wrap(err, "failed to stat file")
	}

	peers, hashes, err := bck.db.loadFile(path)
	if err != nil {
		return errors.Wrap(err, "failed to load file")
	}

	config.Logger.Info("Loaded database", zap.Int("peers", peers), zap.Int("hashes", hashes), zap.Duration("took", time.Now().Sub(start)))

	return nil
}

func (db *Memory) loadFile(filename string) (int, int, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to read file from disk")
	}

	peers, hashes, err := db.decodeBinaryUnsafe(data)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to decode saved data")
	}

	return peers, hashes, nil
}

func (bck *FileBackup) writeFile() (int, error) {
	var encoded []byte
	var err error

	if fast {
		encoded, err = bck.db.encodeBinaryUnsafe()
	} else {
		encoded, err = bck.db.encodeBinaryUnsafeAutoalloc()
	}

	if err != nil {
		return 0, errors.Wrap(err, "failed to encode data")
	}

	if err := ioutil.WriteFile(config.CacheDir+"peers.db", encoded, 0644); err != nil {
		return 0, errors.Wrap(err, "failed to write file to disk")
	}

	return len(encoded), nil
}

// Save encodes and writes the database to a file
func (bck *FileBackup) Save() error {
	config.Logger.Info("Writing database to file")
	start := time.Now()

	size, err := bck.writeFile()
	if err != nil {
		return errors.Wrap(err, "failed to save database")
	}

	config.Logger.Info("Wrote database", zap.Int("size (bytes)", size), zap.Int("hashes", bck.db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))

	return nil
}
