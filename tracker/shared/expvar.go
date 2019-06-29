package shared

var (
	ExpvarAnnounces  int64
	ExpvarScrapes    int64
	ExpvarErrs       int64
	ExpvarClienterrs int64
	ExpvarSeeds      map[[40]byte]bool
	ExpvarLeeches    map[[40]byte]bool
	ExpvarIPs        map[PeerIP]int8
)

func expvarKey(hash, id [20]byte) (result [40]byte) {
	x := append(hash[:], id[:]...)
	copy(result[:], x)
	return
}

func initExpvar() {
	// Might as well alloc capcity at start
	ExpvarSeeds = make(map[[40]byte]bool, 80000)
	ExpvarLeeches = make(map[[40]byte]bool, 80000)
	ExpvarIPs = make(map[PeerIP]int8, 50000)

	if PeerDB == nil {
		panic("peerDB not init before expvars")
	}

	for hash, peermap := range PeerDB {
		for id, peer := range peermap {
			key := expvarKey(hash, id)
			ExpvarIPs[peer.IP]++

			if peer.Complete == true {
				ExpvarSeeds[key] = true
			} else {
				ExpvarLeeches[key] = true
			}
		}
	}
}
