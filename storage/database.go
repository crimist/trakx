/*
	Storage contains all related database interfaces, database types, type pools, and expvar logic.
*/

package storage

import (
	"net/netip"

	"github.com/pkg/errors"

	"github.com/crimist/trakx/config"
)

// Open opens and initializes given database type through config.
func Open() (Database, error) {
	driver, ok := drivers[config.Config.DB.Type]
	if !ok {
		return nil, errors.New("Invalid database driver: '" + config.Config.DB.Type + "'")
	}

	backup, ok := driver.backups[config.Config.DB.Backup.Type]
	if !ok {
		return nil, errors.New("Invalid backup driver: '" + config.Config.DB.Backup.Type + "'")
	}

	if err := driver.db.Init(backup); err != nil {
		return nil, errors.Wrap(err, "failed to init storage driver")
	}

	if err := driver.db.SyncExpvars(); err != nil {
		return nil, errors.Wrap(err, "failed to sync expvars on storage driver")
	}

	return driver.db, nil
}

type Database interface {
	// Used to init the database after open()
	Init(backup Backup) error

	// Internal functions
	Check() bool
	Backup() Backup
	Trim()
	SyncExpvars() error

	Save(netip.Addr, uint16, bool, Hash, PeerID)
	Drop(Hash, PeerID)

	HashStats(Hash) (uint16, uint16)
	PeerList(Hash, uint, bool) [][]byte
	PeerListBytes(Hash, uint) ([]byte, []byte)

	// Number of hashes for stats
	Hashes() int
}

type Backup interface {
	Init(Database) error
	Save() error
	Load() error
}
