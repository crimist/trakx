package shared

var (
	ExpvarAnnounces int64
	ExpvarScrapes   int64
	ExpvarErrs      int64
	ExpvarSeeds   map[[40]byte]bool
	ExpvarLeeches map[[40]byte]bool
	ExpvarIPs     map[string]bool
	ExpvarPeers   map[[40]byte]bool
)

func expvarKey(hash, id [20]byte) (result [40]byte) {
	x := append(hash[:], id[:]...)
	copy(result[:], x)
	return
}

func initExpvar() {
	// Might as well alloc capcity at start
	ExpvarSeeds = make(map[[40]byte]bool, 50000)
	ExpvarLeeches = make(map[[40]byte]bool, 50000)
	ExpvarIPs = make(map[string]bool, 30000)
	ExpvarPeers = make(map[[40]byte]bool, 100000)

	if PeerDB == nil {
		panic("peerDB not init before expvars")
	}

	for hash, peermap := range PeerDB {
		for id, peer := range peermap {
			key := expvarKey(hash, id)
			ExpvarPeers[key] = true
			ExpvarIPs[peer.IP] = true

			if peer.Complete == true {
				ExpvarSeeds[key] = true
			} else {
				ExpvarLeeches[key] = true
			}
		}
	}
}
