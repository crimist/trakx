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

func (d *Database) Load() {
	file, err := os.Open(trackerDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("No database found")
			return
		}
		logger.Panic("db open", zap.Error(err))
	}
	decoder := gob.NewDecoder(file)

	if err := decoder.Decode(&db); err != nil {
		logger.Panic("db gob decoder", zap.Error(err))
	}

	logger.Info("Loaded database", zap.Int("hashes", len(db)))
}

// Write dumps the database to a file
func (d *Database) Write() {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	d.Clean()

	if err := encoder.Encode(&db); err != nil {
		logger.Error("db gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(trackerDBFilename, buff.Bytes(), 0644); err != nil {
		logger.Error("db writefile", zap.Error(err))
	}

	logger.Info("Wrote database", zap.Int("hashes", len(db)))
}
