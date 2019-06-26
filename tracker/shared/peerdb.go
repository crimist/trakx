package shared

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap"
)

type PeerDatabase map[Hash]map[PeerID]Peer

// Clean removes all peers that haven't checked in in CleanTimeout
func (d *PeerDatabase) Clean() {
	for hash, peermap := range PeerDB {
		for id, peer := range peermap {
			if peer.LastSeen < time.Now().Unix()-int64(CleanTimeout) {
				peer.Delete(hash, id)
			}
		}
		if len(peermap) == 0 {
			delete(PeerDB, hash)
		}
	}

	Logger.Info("Cleaned database")
}

func (d *PeerDatabase) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := gob.NewDecoder(file)
	return decoder.Decode(&PeerDB)
}

// Load loads a database into memory
func (d *PeerDatabase) Load() {
	loadtemp := false

	infoFull, err := os.Stat(PeerDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("No full database")
			loadtemp = true
		} else {
			Logger.Error("os.Stat", zap.Error(err))
		}
	}
	infoTemp, err := os.Stat(PeerDBFilename + ".tmp")
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
		if err := d.load(PeerDBTempFilename); err != nil {
			Logger.Info("Loading temp db failed", zap.Error(err))

			if err := d.load(PeerDBFilename); err != nil {
				Logger.Info("Loading full db failed", zap.Error(err))
				return
			} else {
				loaded = "full"
			}
		} else {
			loaded = "temp"
		}
	} else {
		if err := d.load(PeerDBFilename); err != nil {
			Logger.Info("Loading full db failed", zap.Error(err))

			if err := d.load(PeerDBTempFilename); err != nil {
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
func (d *PeerDatabase) Write(istemp bool) {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	d.Clean() // Clean to remove any nil refs

	if err := encoder.Encode(&PeerDB); err != nil {
		Logger.Error("db gob encoder", zap.Error(err))
	}

	filename := PeerDBFilename
	if istemp {
		filename = PeerDBTempFilename
	}
	if err := ioutil.WriteFile(filename, buff.Bytes(), 0644); err != nil {
		Logger.Error("db writefile", zap.Error(err))
	}

	Logger.Info(fmt.Sprintf("Wrote database %v", filename), zap.Int("hashes", len(PeerDB)))
}
