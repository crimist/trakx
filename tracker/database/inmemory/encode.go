package inmemory

import (
	"archive/zip"
	"bytes"
	"encoding/gob"
	"encoding/hex"

	"go.uber.org/zap"
)

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
