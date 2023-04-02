/*
	Storage contains all related database interfaces, database types, type pools, and expvar logic.
*/

package storage

import (
	"net/netip"

	"github.com/pkg/errors"
)

type Config struct {
}

// Open initalizes a database using the given driver
func Open(driver string, config Config) (Database, error) {
	creator, ok := drivers[driver]
	if !ok {
		return nil, errors.New("database driver does not exist with name '" + driver + "'")
	}

	db, err := creator(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create database of type '"+driver+"'")
	}

	return db, nil
}

type Database interface {
	PeerAdd(hash Hash, peerID PeerID, addr netip.Addr, port uint16, complete bool)
	PeerRemove(hash Hash, peerID PeerID)

	// TODO: figure out the parameters for these methods
	TorrentStats(Hash) (seeds uint16, leeches uint16)
	TorrentPeers(Hash, uint, bool) [][]byte
	TorrentPeersBytes(Hash, uint) ([]byte, []byte)

	// Hashcount returns the total number of hashes registered in the database
	HashCount() (hashes int)
}
