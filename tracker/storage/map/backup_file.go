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

// FileBackup backs up the peer database to a local file.
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

	_, err := os.Stat(config.Conf.DB.Backup.Path)
	if err != nil {
		// If the file doesn't exist than create an empty database and return success
		if os.IsNotExist(err) {
			bck.db.make()
			config.Logger.Info("Database file not found, created empty database", zap.String("filepath", config.Conf.DB.Backup.Path))
			return nil
		}

		return errors.Wrap(err, "failed to stat file")
	}

	peers, hashes, err := bck.db.loadFile(config.Conf.DB.Backup.Path)
	if err != nil {
		return errors.Wrap(err, "failed to load file")
	}

	config.Logger.Info("Loaded database", zap.Duration("time", time.Since(start)), zap.Int("peers", peers), zap.Int("hashes", hashes))

	return nil
}

func (db *Memory) loadFile(filename string) (peers int, hashes int, err error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		err = errors.Wrap(err, "failed to read file from disk")
		return
	}

	peers, hashes, err = db.decodeBinary(data)
	err = errors.Wrap(err, "failed to decode saved data")

	return
}

func (bck *FileBackup) writeFile() (int, error) {
	var encoded []byte
	var err error

	encoded, err = bck.db.encodeBinary()

	if err != nil {
		return 0, errors.Wrap(err, "failed to encode db")
	}

	if err := ioutil.WriteFile(config.Conf.DB.Backup.Path, encoded, 0644); err != nil {
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

	config.Logger.Info("Wrote database", zap.Int("size (bytes)", size), zap.Int("hashes", bck.db.Hashes()), zap.Duration("duration", time.Since(start)))

	return nil
}
