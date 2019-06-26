package shared

var (
	ExpvarAnnounces int64
	ExpvarScrapes   int64
	ExpvarErrs      int64

	// !x test
	ExpvarSeeds   map[PeerID]bool
	ExpvarLeeches map[PeerID]bool
	ExpvarIPs     map[string]bool
	ExpvarPeers   map[PeerID]bool
)

// !x
func initExpvar() {
	// Might as well alloc capcity at start
	ExpvarSeeds = make(map[PeerID]bool, 50000)
	ExpvarLeeches = make(map[PeerID]bool, 50000)
	ExpvarIPs = make(map[string]bool, 30000)
	ExpvarPeers = make(map[PeerID]bool, 100000)

	if PeerDB == nil {
		panic("peerDB not init before expvars")
	}

	for _, peermap := range PeerDB {
		for id, peer := range peermap {
			ExpvarPeers[id] = true
			ExpvarIPs[peer.IP] = true

			if peer.Complete == true {
				ExpvarSeeds[id] = true
			} else {
				ExpvarLeeches[id] = true
			}
		}
	}
}
