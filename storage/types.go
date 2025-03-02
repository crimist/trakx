package storage

import "net/netip"

type IPVersion uint8

const (
	IPv4 IPVersion = 1 << iota
	IPv6
)

type (
	// Hash stores a BitTorrent infohash.
	Hash [20]byte
	// PeerID stores a BitTorrent peer ID.
	PeerID [20]byte

	// Peer contains requied peer information for database.
	Peer struct {
		Complete bool
		IP       netip.Addr
		Port     uint16
		LastSeen int64
	}
)
