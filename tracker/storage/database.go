/*
	Storage contains all related database interfaces, database types, type pools, and expvar logic.
*/

package storage

import (
	"github.com/pkg/errors"

	"github.com/crimist/trakx/tracker/config"
)

// Open opens and initializes given database type through config.
func Open() (Database, error) {
	driver, ok := drivers[config.Conf.Database.Type]
	if !ok {
		return nil, errors.New("Invalid database driver: '" + config.Conf.Database.Type + "'")
	}

	backup, ok := driver.backups[config.Conf.Database.Backup]
	if !ok {
		return nil, errors.New("Invalid backup driver: '" + config.Conf.Database.Backup + "'")
	}

	if err := driver.db.Init(backup); err != nil {
		return nil, errors.Wrap(err, "failed to init storage driver")
	}

	if err := driver.db.Expvar(); err != nil {
		return nil, errors.Wrap(err, "failed to init expvar on storage driver")
	}

	// set peerlistMax based on max numwant limit
	peerlistMax = 6 * int(config.Conf.Tracker.Numwant.Limit)

	return driver.db, nil
}

type Database interface {
	// Used to init the database after open()
	Init(backup Backup) error

	// Internal functions
	Check() bool
	Backup() Backup
	Trim()
	Expvar() error

	Save(PeerIP, uint16, bool, Hash, PeerID)
	Drop(Hash, PeerID)

	HashStats(Hash) (uint16, uint16)
	PeerList(Hash, int, bool) [][]byte
	PeerListBytes(Hash, int) *Peerlist

	// Only used for expvar
	Hashes() int
}

type Backup interface {
	Init(Database) error
	Save() error
	Load() error
}
