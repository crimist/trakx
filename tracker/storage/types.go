package storage

type (
	// Hash stores a BitTorrent infohash.
	Hash [20]byte
	// PeerID stores a BitTorrent peer ID.
	PeerID [20]byte

	// Peer contains requied peer information for database.
	Peer struct {
		Complete bool
		IP       PeerIP
		Port     uint16
		LastSeen int64
	}
)
