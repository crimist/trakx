package tracker

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap"
)

type Database map[Hash]map[PeerID]Peer

func (d *Database) Clean() {
	for hash, peermap := range db {
		for id, peer := range peermap {
			if peer.LastSeen < time.Now().Unix()-int64(trackerCleanTimeout) {
				delete(peermap, id)
				expvarCleaned++
			}
		}
		if len(peermap) == 0 {
			delete(db, hash)
		}
	}
}

func (d *Database) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&db)
}

// Load loads a database into memory
func (d *Database) Load() {
	loadtemp := false

	infoFull, err := os.Stat(trackerDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("No full database")
			loadtemp = true
		} else {
			logger.Panic("os.Stat", zap.Error(err))
		}
	}
	infoTemp, err := os.Stat(trackerDBFilename + ".tmp")
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("No temp database")
			if loadtemp {
				logger.Info("No database found")
				return
			}
		} else {
			logger.Panic("os.Stat", zap.Error(err))
		}
	}

	if !loadtemp {
		if infoTemp.ModTime().UnixNano() > infoFull.ModTime().UnixNano() {
			loadtemp = true
		}
	}

	if loadtemp == true {
		if err := d.load(trackerDBTempFilename); err != nil {
			logger.Info("Loading temp db failed", zap.Error(err))

			if err := d.load(trackerDBFilename); err != nil {
				logger.Info("Loading full db failed", zap.Error(err))
				return
			}
		}
	} else {
		if err := d.load(trackerDBFilename); err != nil {
			logger.Info("Loading full db failed", zap.Error(err))

			if err := d.load(trackerDBTempFilename); err != nil {
				logger.Info("Loading temp db failed", zap.Error(err))
				return
			}
		}
	}

	logger.Info("Loaded database", zap.Int("hashes", len(db)))
}

// Write dumps the database to a file
func (d *Database) Write(istemp bool) {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	d.Clean() // Clean to remove any nil refs

	if err := encoder.Encode(&db); err != nil {
		logger.Error("db gob encoder", zap.Error(err))
	}

	filename := trackerDBFilename
	if istemp {
		filename += ".tmp"
	}
	if err := ioutil.WriteFile(filename, buff.Bytes(), 0644); err != nil {
		logger.Error("db writefile", zap.Error(err))
	}

	logger.Info("Wrote database", zap.Int("hashes", len(db)))
}
