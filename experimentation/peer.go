package experimentation

type Hash [20]byte
type PeerID [20]byte
type PeerIP [4]byte

type Peer struct {
	Complete bool
	IP       PeerIP
	Port     uint16
	LastSeen int64
}
