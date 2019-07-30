package shared

type PeerID [20]byte
type PeerIP [4]byte

// Peer holds peer information stores in the database
type Peer struct {
	IP       PeerIP
	Port     uint16
	Complete bool
	LastSeen int64
}
