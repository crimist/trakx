package gomap

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/crimist/trakx/tracker/storage"
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
	bck.db.logger.Info("Loading database from file")
	start := time.Now()

	_, err := os.Stat(bck.db.conf.Database.Peer.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("Database file not found: " + err.Error())
		}

		return errors.New("os.Stat failed: " + err.Error())
	}

	if err := bck.db.loadFile(bck.db.conf.Database.Peer.Filename); err != nil {
		return errors.New("Failed to load file: " + err.Error())
	}

	bck.db.logger.Info("Loaded database", zap.Int("hashes", bck.db.Hashes()), zap.Duration("took", time.Now().Sub(start)))

	return nil
}

func (db *Memory) loadFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := db.decode(data); err != nil {
		return err
	}

	return nil
}

// like trim() this uses costly locking but it's worth it to prevent blocking
func (bck *FileBackup) writeFile() (float32, error) {
	filename := bck.db.conf.Database.Peer.Filename

	encoded, err := bck.db.encode()
	if err != nil {
		return 0, err
	}

	size := float32(len(encoded) / 1024.0 / 1024.0)
	if err := ioutil.WriteFile(filename, encoded, 0644); err != nil {
		return 0, errors.New("Failed to write file: " + err.Error())
	}
	return size, nil
}

func (bck *FileBackup) Save() error {
	bck.db.logger.Info("Writing database to file")
	start := time.Now()
	size, err := bck.writeFile()

	if err != nil {
		bck.db.logger.Info("Failed to write database", zap.Duration("duration", time.Now().Sub(start)))
	} else {
		bck.db.logger.Info("Wrote database", zap.Float32("size", size), zap.Int("hashes", bck.db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
	}

	return err
}
