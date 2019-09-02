package database

import "github.com/syc0x00/trakx/tracker/shared"

type Database interface {
	Check() bool
	Backup() Backup
	Trim()
	InitExpvar()

	Save(*shared.Peer, *shared.Hash, *shared.PeerID)
	Drop(*shared.Peer, *shared.Hash, *shared.PeerID)

	Hashes() int
	HashStats(*shared.Hash) (int32, int32)
	PeerList(*shared.Hash, int, bool) []string
	PeerListBytes(*shared.Hash, int) []byte
}

type Backup interface {
	Init(Database) error
	SaveTmp() error
	SaveFull() error
	Load() error
}
