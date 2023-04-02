/*
	Storage contains all related database interfaces, database types, type pools, and expvar logic.
*/

package storage

import (
	"net/netip"

	"github.com/pkg/errors"
)

type DatabaseConfig struct {
}

// OpenPeerDatabase opens and initializes given database type through config.
func OpenPeerDatabase(databaseDriver string, backupDriver string, config DatabaseConfig) (Database, error) {
	driver, ok := databaseDrivers[databaseDriver]
	if !ok {
		return nil, errors.New("database driver does not exist with name: '" + databaseDriver + "'")
	}

	backup, ok := driver.backupDrivers[backupDriver]
	if !ok {
		return nil, errors.New("database backup does not exist with name: '" + backupDriver + "'")
	}

	if err := driver.database.Initialize(backup, config); err != nil {
		return nil, errors.Wrap(err, "failed to init storage driver")
	}

	if err := driver.database.SyncExpvars(); err != nil {
		return nil, errors.Wrap(err, "failed to sync expvars on storage driver")
	}

	return driver.database, nil
}

// TODO: refacor these method names, create different methods depending on IP version

type Database interface {
	// init when opened
	Initialize(persistanceDriver Persistance, config DatabaseConfig) error

	// Internal functions
	PersistToDisk() error
	Trim()
	SyncExpvars() error

	Save(netip.Addr, uint16, bool, Hash, PeerID)
	Drop(Hash, PeerID)

	TorrentStats(Hash) (seeds uint16, leeches uint16)
	PeerList(Hash, uint, bool) [][]byte
	PeerListBytes(Hash, uint) ([]byte, []byte)

	HashCount() (hashes int)
}

type Persistance interface {
	Init(Database) error
	WriteToDisk() error
	ReadFromDisk() error
}
