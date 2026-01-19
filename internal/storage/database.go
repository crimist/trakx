/*
	Storage contains all related database interfaces, database types, type pools, and expvar logic.
*/

package storage

import (
	"io"
	"net/netip"
)

// Snapshotter defines the interface for components that can be snapshotted and restored.
type Snapshotter interface {
	// Snapshot writes the component's state to the provided writer
	Snapshot(w io.Writer) error
	// Restore reads the component's state from the provided reader
	Restore(r io.Reader) error
}

type Database interface {
	PeerAdd(hash Hash, peerID PeerID, addr netip.Addr, port uint16, complete bool)
	PeerRemove(hash Hash, peerID PeerID)

	TorrentStats(hash Hash) (seeds uint16, leeches uint16)
	TorrentPeers(hash Hash, numWant uint, includePeerID bool) [][]byte
	TorrentPeersCompact(hash Hash, numWant uint, wantedIPs IPVersion) PeerLists

	// Torrents returns the total number of torrents registered in the database
	Torrents() (numtorrents int)
}
