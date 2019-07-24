package shared

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"time"

	"go.uber.org/zap"
)

type PeerDatabase map[Hash]map[PeerID]Peer

// Clean removes all peers that haven't checked in since timeout
func (d *PeerDatabase) Clean() {
	var peers, hashes int
	now := time.Now().Unix()

	for hash, peermap := range PeerDB {
		for id, peer := range peermap {
			if now-peer.LastSeen > Config.DBCleanTimeout {
				peer.Delete(hash, id)
				peers++
			}
		}
		if len(peermap) == 0 {
			delete(PeerDB, hash)
			hashes++
		}
	}

	Logger.Info("Cleaned PeerDatabase", zap.Int("peers", peers), zap.Int("hashes", hashes))
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

	infoFull, err := os.Stat(Config.Database.Peer)
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("No full peerdb")
			loadtemp = true
		} else {
			Logger.Error("os.Stat", zap.Error(err))
		}
	}
	infoTemp, err := os.Stat(Config.Database.Peer + ".tmp")
	if err != nil {
		if os.IsNotExist(err) {
			Logger.Info("No temp peerdb")
			if loadtemp {
				Logger.Info("No peerdb found")
				PeerDB = make(PeerDatabase)
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
		if err := d.load(Config.Database.Peer + ".tmp"); err != nil {
			Logger.Info("Loading temp peerdb failed", zap.Error(err))

			if err := d.load(Config.Database.Peer); err != nil {
				Logger.Info("Loading full peerdb failed", zap.Error(err))
				PeerDB = make(PeerDatabase)
				return
			} else {
				loaded = "full"
			}
		} else {
			loaded = "temp"
		}
	} else {
		if err := d.load(Config.Database.Peer); err != nil {
			Logger.Info("Loading full peerdb failed", zap.Error(err))

			if err := d.load(Config.Database.Peer + ".tmp"); err != nil {
				Logger.Info("Loading temp peerdb failed", zap.Error(err))
				PeerDB = make(PeerDatabase)
				return
			} else {
				loaded = "temp"
			}
		} else {
			loaded = "full"
		}
	}

	Logger.Info("Loaded peerdb", zap.String("type", loaded), zap.Int("hashes", len(PeerDB)))
}

// WriteTmp dumps the database to the tmp file
func (d *PeerDatabase) WriteTmp() {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(&PeerDB); err != nil {
		Logger.Error("peerdb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(Config.Database.Peer+".tmp", buff.Bytes(), 0644); err != nil {
		Logger.Error("peerdb writefile", zap.Error(err))
	}

	Logger.Info("Wrote temp peerdb", zap.Int("hashes", len(PeerDB)))
}

// WriteFull dumps the database to the db file
func (d *PeerDatabase) WriteFull() {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	d.Clean() // Clean to remove any nil refs

	if err := encoder.Encode(&PeerDB); err != nil {
		Logger.Error("peerdb gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(Config.Database.Peer, buff.Bytes(), 0644); err != nil {
		Logger.Error("peerdb writefile", zap.Error(err))
	}

	Logger.Info("Wrote full peerdb", zap.Int("hashes", len(PeerDB)))
}
