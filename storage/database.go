/*
	Storage contains all related database interfaces, database types, type pools, and expvar logic.
*/

package storage

import (
	"net/netip"
)

type Database interface {
	PeerAdd(hash Hash, peerID PeerID, addr netip.Addr, port uint16, complete bool)
	PeerRemove(hash Hash, peerID PeerID)

	TorrentStats(hash Hash) (seeds uint16, leeches uint16)
	TorrentPeers(hash Hash, numWant uint, includePeerID bool) [][]byte
	TorrentPeersCompact(hash Hash, numWant uint, wantedIPs IPVersion) (peers4 []byte, peers6 []byte)

	// Torrents returns the total number of torrents registered in the database
	Torrents() (numtorrents int)
}
