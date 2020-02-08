package storage

import (
	"github.com/pkg/errors"

	"github.com/crimist/trakx/tracker/shared"
)

func Open(conf *shared.Config) (Database, error) {
	driver, ok := drivers[conf.Database.Type]
	if !ok {
		return nil, errors.New("Invalid database driver: '" + conf.Database.Type + "'")
	}

	backup, ok := driver.backups[conf.Database.Backup]
	if !ok {
		return nil, errors.New("Invalid backup driver: '" + conf.Database.Backup + "'")
	}

	if err := driver.db.Init(conf, backup); err != nil {
		return nil, errors.Wrap(err, "failed to init storage driver")
	}

	if err := driver.db.Expvar(); err != nil {
		return nil, errors.Wrap(err, "failed to init expvar on storage driver")
	}

	*peerlistMax = 6 * int(conf.Tracker.Numwant.Limit)

	return driver.db, nil
}

type Database interface {
	// Used to init the database after open()
	Init(conf *shared.Config, backup Backup) error

	// Internal functions
	Check() bool
	Backup() Backup
	Trim()
	Expvar() error

	Save(*Peer, Hash, PeerID)
	Drop(Hash, PeerID)

	HashStats(Hash) (int32, int32)
	PeerList(Hash, int, bool) []string
	PeerListBytes(Hash, int) *Peerlist

	// Only used for expvar
	Hashes() int
}

type Backup interface {
	Init(Database) error
	Save() error
	Load() error
}
