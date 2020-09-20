package storage

type (
	Hash   [20]byte
	PeerID [20]byte

	// Peer holds peer information stores in the database
	Peer struct {
		Complete bool
		IP       PeerIP
		Port     uint16
		LastSeen int64
	}
)
