package storage

// PeerLists holds compact peer lists and their release hook.
type PeerLists struct {
	V4      []byte
	V6      []byte
	release func()
}

// NewPeerLists constructs a PeerLists with a release hook.
func NewPeerLists(v4, v6 []byte, release func()) PeerLists {
	return PeerLists{
		V4:      v4,
		V6:      v6,
		release: release,
	}
}

// Release returns the underlying buffers to their pool.
func (p *PeerLists) Release() {
	if p.release == nil {
		return
	}
	p.release()
	p.release = nil
}
