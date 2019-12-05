package storage

import (
	"errors"

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
		return nil, errors.New("Failed to initialize storage driver: " + err.Error())
	}

	if err := driver.db.Expvar(); err != nil {
		return nil, errors.New("Expvar() call failed: " + err.Error())
	}

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

	Save(*Peer, *Hash, *PeerID)
	Drop(*Peer, *Hash, *PeerID)

	HashStats(*Hash) (int32, int32)
	PeerList(*Hash, int, bool) []string
	PeerListBytes(*Hash, int) []byte

	// Only used for expvar
	Hashes() int
}

type Backup interface {
	Init(Database) error
	Save() error
	Load() error
}
