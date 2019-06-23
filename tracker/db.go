package tracker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap"
)

type Database map[Hash]map[PeerID]Peer

// Clean removes all peers that haven't checked in in trackerCleanTimeout
func (d *Database) Clean() {
	expvarCleanedHashes = 0
	expvarCleanedPeers = 0

	for hash, peermap := range PeerDB {
		for id, peer := range peermap {
			if peer.LastSeen < time.Now().Unix()-int64(trackerCleanTimeout) {
				delete(peermap, id)
				expvarCleanedPeers++
			}
		}
		if len(peermap) == 0 {
			delete(PeerDB, hash)
			expvarCleanedHashes++
		}
	}

	Logger.Info("Cleaned database", zap.Int64("peers", expvarCleanedPeers), zap.Int64("Hashes", expvarCleanedHashes))
}

func (d *Database) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&PeerDB)
}

// Load loads a database into memory
func (d *Database) Load() {
	loadtemp := false

	infoFull, err := os.Stat(trackerDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("No full database")
			loadtemp = true
		} else {
			Logger.Error("os.Stat", zap.Error(err))
		}
	}
	infoTemp, err := os.Stat(trackerDBFilename + ".tmp")
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("No temp database")
			if loadtemp {
				Logger.Info("No database found")
				return
			}
		} else {
			Logger.Error("os.Stat", zap.Error(err))
		}
	}

	if infoFull != nil && infoTemp != nil {
		if infoTemp.ModTime().UnixNano() > infoFull.ModTime().UnixNano() {
			loadtemp = true
		}
	}

	loaded := ""
	if loadtemp == true {
		if err := d.load(trackerDBTempFilename); err != nil {
			Logger.Info("Loading temp db failed", zap.Error(err))

			if err := d.load(trackerDBFilename); err != nil {
				Logger.Info("Loading full db failed", zap.Error(err))
				return
			} else {
				loaded = "full"
			}
		} else {
			loaded = "temp"
		}
	} else {
		if err := d.load(trackerDBFilename); err != nil {
			Logger.Info("Loading full db failed", zap.Error(err))

			if err := d.load(trackerDBTempFilename); err != nil {
				Logger.Info("Loading temp db failed", zap.Error(err))
				return
			} else {
				loaded = "temp"
			}
		} else {
			loaded = "full"
		}
	}

	Logger.Info(fmt.Sprintf("Loaded %v database", loaded), zap.Int("hashes", len(PeerDB)))
}

// Write dumps the database to a file
func (d *Database) Write(istemp bool) {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	d.Clean() // Clean to remove any nil refs

	if err := encoder.Encode(&PeerDB); err != nil {
		Logger.Error("db gob encoder", zap.Error(err))
	}

	filename := trackerDBFilename
	if istemp {
		filename = trackerDBTempFilename
	}
	if err := ioutil.WriteFile(filename, buff.Bytes(), 0644); err != nil {
		Logger.Error("db writefile", zap.Error(err))
	}

	Logger.Info(fmt.Sprintf("Wrote database %v", filename), zap.Int("hashes", len(PeerDB)))
}
