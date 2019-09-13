package gomap

import (
	"archive/zip"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/syc0x00/trakx/tracker/storage"
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
	bck.db.logger.Info("Loading database")
	start := time.Now()
	loadtemp := false

	infoFull, err := os.Stat(bck.db.conf.Database.Peer.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			bck.db.logger.Info("No full peerdb")
			loadtemp = true
		} else {
			bck.db.logger.Error("os.Stat", zap.Error(err))
		}
	}
	infoTemp, err := os.Stat(bck.db.conf.Database.Peer.Filename + ".tmp")
	if err != nil {
		if os.IsNotExist(err) {
			bck.db.logger.Info("No temp peerdb")
			if loadtemp {
				bck.db.logger.Info("No peerdb found")
				bck.db.make()
				return nil
			}
		} else {
			bck.db.logger.Error("os.Stat", zap.Error(err))
		}
	}

	if infoFull != nil && infoTemp != nil {
		if infoTemp.ModTime().UnixNano() > infoFull.ModTime().UnixNano() {
			loadtemp = true
		}
	}

	loaded := ""
	if loadtemp == true {
		if err := bck.db.loadFile(bck.db.conf.Database.Peer.Filename + ".tmp"); err != nil {
			bck.db.logger.Info("Loading temp peerdb failed", zap.Error(err))

			if err := bck.db.loadFile(bck.db.conf.Database.Peer.Filename); err != nil {
				bck.db.logger.Info("Loading full peerdb failed", zap.Error(err))
				bck.db.make()
				return errors.New("Failed to load any db")
			} else {
				loaded = "full"
			}
		} else {
			loaded = "temp"
		}
	} else {
		if err := bck.db.loadFile(bck.db.conf.Database.Peer.Filename); err != nil {
			bck.db.logger.Info("Loading full peerdb failed", zap.Error(err))

			if err := bck.db.loadFile(bck.db.conf.Database.Peer.Filename + ".tmp"); err != nil {
				bck.db.logger.Info("Loading temp peerdb failed", zap.Error(err))
				bck.db.make()
				return errors.New("Failed to load any db")
			} else {
				loaded = "temp"
			}
		} else {
			loaded = "full"
		}
	}

	bck.db.logger.Info("Loaded database", zap.String("type", loaded), zap.Int("hashes", bck.db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))

	return nil
}

func (db *Memory) loadFile(filename string) error {
	var hash storage.Hash
	db.make()

	archive, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		hashbytes, err := hex.DecodeString(file.Name)
		if err != nil {
			return err
		}
		copy(hash[:], hashbytes)
		peermap := db.makePeermap(&hash)

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

// like trim() this uses costly locking but it's worth it to prevent blocking
func (bck *FileBackup) writeToFile(temp bool) error {
	filename := bck.db.conf.Database.Peer.Filename
	if temp {
		filename += ".tmp"
	} else {
		bck.db.Trim()
	}

	encoded, err := bck.db.encode()
	if err != nil {
		return err
	}

	bck.db.logger.Info("Writing zip to file", zap.Float32("mb", float32(len(encoded)/1024.0/1024.0)))
	if err := ioutil.WriteFile(filename, encoded, 0644); err != nil {
		bck.db.logger.Error("Database writefile failed", zap.Error(err))
		return errors.New("Failed to write file")
	}
	return nil
}

// SaveTmp writes the database to tmp file
func (bck *FileBackup) SaveTmp() error {
	bck.db.logger.Info("Writing temp database")
	start := time.Now()
	err := bck.writeToFile(true)

	if err != nil {
		bck.db.logger.Info("Failed to write temp database", zap.Duration("duration", time.Now().Sub(start)))
	} else {
		bck.db.logger.Info("Wrote temp database", zap.Int("hashes", bck.db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
	}

	return err
}

// SaveFull writes the database to file
func (bck *FileBackup) SaveFull() error {
	bck.db.logger.Info("Writing full database")
	start := time.Now()
	err := bck.writeToFile(false)

	if err != nil {
		bck.db.logger.Info("Failed to write full database", zap.Duration("duration", time.Now().Sub(start)))
	} else {
		bck.db.logger.Info("Wrote full database", zap.Int("hashes", bck.db.Hashes()), zap.Duration("duration", time.Now().Sub(start)))
	}

	return err
}
