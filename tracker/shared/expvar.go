package shared

var (
	ExpvarAnnounces  int64
	ExpvarScrapes    int64
	ExpvarErrs       int64
	ExpvarClienterrs int64
	ExpvarSeeds      int64
	ExpvarLeeches    int64
	ExpvarIPs        map[PeerIP]int8
)

func initExpvar() {
	// Might as well alloc capcity at start
	ExpvarIPs = make(map[PeerIP]int8, 50000)

	if PeerDB == nil {
		panic("peerDB not init before expvars")
	}

	for _, peermap := range PeerDB {
		for _, peer := range peermap {
			ExpvarIPs[peer.IP]++
			if peer.Complete == true {
				ExpvarSeeds++
			} else {
				ExpvarLeeches++
			}
		}
	}
}
