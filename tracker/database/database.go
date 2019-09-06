package database

import (
	"errors"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

func Open(name, backupname string) (Database, Backup, error) {
	driver, ok := drivers[name]
	if !ok {
		return nil, nil, errors.New("Invalid database type: '" + name + "'")
	}
	backup, ok := driver.backups[backupname]
	if !ok {
		return nil, nil, errors.New("Invalid backup type: '" + backupname + "'")
	}
	return driver.db, backup, nil
}

type Database interface {
	// Used to init the database after open()
	Init(conf *shared.Config, logger *zap.Logger, backup Backup)

	// Internal functions
	Check() bool
	Backup() Backup
	Trim()
	Expvar()

	Save(*shared.Peer, *shared.Hash, *shared.PeerID)
	Drop(*shared.Peer, *shared.Hash, *shared.PeerID)

	HashStats(*shared.Hash) (int32, int32)
	PeerList(*shared.Hash, int, bool) []string
	PeerListBytes(*shared.Hash, int) []byte

	// Only used for expvar
	Hashes() int
}

type Backup interface {
	Init(Database) error
	SaveTmp() error
	SaveFull() error
	Load() error
}
