package database

import (
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const defaultFilePermission = 0640 // rw-r-----

func WriteSnapshotFile(db *Database, path string) (err error) {
	zap.L().Info("Persisting database to file", zap.String("path", path))
	start := time.Now()

	tmpDir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(tmpDir, "trakx-db-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp db file")
	}
	tmpPath := tmpFile.Name()
	defer func() {
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	if err = db.Snapshot(tmpFile); err != nil {
		return errors.Wrap(err, "failed to write database snapshot")
	}
	if err = tmpFile.Sync(); err != nil {
		return errors.Wrap(err, "failed to sync database snapshot")
	}
	if err = tmpFile.Close(); err != nil {
		return errors.Wrap(err, "failed to close database snapshot")
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return errors.Wrap(err, "failed to replace database snapshot")
	}
	if err = os.Chmod(path, defaultFilePermission); err != nil {
		return errors.Wrap(err, "failed to set database snapshot permissions")
	}

	zap.L().Info("Persisted database to file", zap.Duration("elapsed", time.Since(start)))
	return nil
}

func ReadSnapshotFile(db *Database, path string) error {
	zap.L().Info("Loading database from file", zap.String("path", path))
	start := time.Now()

	file, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "failed to open database snapshot")
	}
	defer file.Close()

	if err := db.Restore(file); err != nil {
		return errors.Wrap(err, "failed to restore database snapshot")
	}

	zap.L().Info("Loaded database from file", zap.Int("torrents", db.Torrents()), zap.Duration("elapsed", time.Since(start)))
	return nil
}
