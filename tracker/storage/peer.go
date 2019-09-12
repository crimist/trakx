package storage

type PeerID [20]byte
type PeerIP [4]byte

// Peer holds peer information stores in the database
type Peer struct {
	Complete bool
	IP       PeerIP
	Port     uint16
	LastSeen int64
}
