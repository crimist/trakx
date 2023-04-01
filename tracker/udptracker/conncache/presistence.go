package conncache

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"os"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const fileMode = 0644

func encodeCacheEntries(connCache *ConnectionCache) ([]byte, error) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	connCache.mutex.Lock()
	gob.NewEncoder(writer).Encode(connCache.entries)
	connCache.mutex.Unlock()

	if err := writer.Flush(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func PersistEntriesToFile(path string, connCache *ConnectionCache) error {
	zap.L().Info("persisting connection cache to file")
	start := time.Now()

	encoded, err := encodeCacheEntries(connCache)
	if err != nil {
		return errors.Wrap(err, "failed to encode connection cache")
	}

	if err := os.WriteFile(path, encoded, fileMode); err != nil {
		return errors.Wrap(err, "failed to write file")
	}

	zap.L().Info("persisted connection cache entries to file", zap.String("path", path), zap.Int("entrycount", connCache.EntryCount()), zap.Duration("elapsed", time.Since(start)))
	return nil

}

func decodeCacheEntries(data []byte) (entryMap, error) {
	// len(data)/18 is a rough estimate of the number of entries
	entries := make(entryMap, len(data)/18)
	reader := bufio.NewReader(bytes.NewBuffer(data))
	err := gob.NewDecoder(reader).Decode(&entries)

	return entries, err
}

func loadEntriesFromFile(path string) (entryMap, error) {
	zap.L().Info("loading connection cache entries from file")
	start := time.Now()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read connection database file from disk")
	}

	entries, err := decodeCacheEntries(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshall binary data")

	}

	zap.L().Info("loadad connection cache entries from file", zap.String("path", path), zap.Int("entrycount", len(entries)), zap.Duration("elapsed", time.Since(start)))
	return entries, nil
}
