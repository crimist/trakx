package database

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"io/ioutil"
	"os"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type MemoryBackup struct {
	db *Memory
}

func (bck MemoryBackup) Save() int {
	encoded := bck.db.encode()
	if encoded == nil {
		return -1
	}

	bck.db.logger.Info("Writing zip to file", zap.Float32("mb", float32(len(encoded)/1024.0/1024.0)))
	if err := ioutil.WriteFile(bck.db.conf.Database.Peer.Filename, encoded, 0644); err != nil {
		bck.db.logger.Error("Database writefile failed", zap.Error(err))
		return -1
	}

	return len(encoded)
}

func (db *Memory) make() {
	db.hashmap = make(map[shared.Hash]*subPeerMap, initCap)
}

func (bck MemoryBackup) Load() error {
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
				return nil
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
				return nil
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
	var hash shared.Hash
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

func (db *Memory) encode() []byte {
	var buff bytes.Buffer
	archive := zip.NewWriter(&buff)

	db.mu.RLock()
	for hash, submap := range db.hashmap {
		db.mu.RUnlock()

		submap.RLock()
		writer, err := archive.Create(hex.EncodeToString(hash[:]))
		if err != nil {
			db.logger.Error("Failed to create in archive", zap.Error(err), zap.Any("hash", hash[:]))
			submap.RUnlock()
			continue
		}
		if err := gob.NewEncoder(writer).Encode(submap.peers); err != nil {
			db.logger.Warn("Failed to encode a peermap", zap.Error(err), zap.Any("hash", hash[:]))
		}
		submap.RUnlock()

		db.mu.RLock()
	}
	db.mu.RUnlock()

	if err := archive.Close(); err != nil {
		db.logger.Error("Failed to close archive", zap.Error(err))
		return nil
	}

	return buff.Bytes()
}
