package shared

var (
	ExpvarAnnounces  int64
	ExpvarScrapes    int64
	ExpvarErrs       int64
	ExpvarClienterrs int64
	ExpvarSeeds      map[[40]byte]bool
	ExpvarLeeches    map[[40]byte]bool
	ExpvarIPs        map[string]int8
	ExpvarPeers      map[[40]byte]bool // could just add seeds & leeches to get this. more efficient
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
	ExpvarIPs = make(map[string]int8, 30000)
	ExpvarPeers = make(map[[40]byte]bool, 100000)

	if PeerDB == nil {
		panic("peerDB not init before expvars")
	}

	for hash, peermap := range PeerDB {
		for id, peer := range peermap {
			key := expvarKey(hash, id)
			ExpvarPeers[key] = true
			ExpvarIPs[peer.IP]++

			if peer.Complete == true {
				ExpvarSeeds[key] = true
			} else {
				ExpvarLeeches[key] = true
			}
		}
	}
}
