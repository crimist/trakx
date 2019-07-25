package shared

var (
	ExpvarConnects    int64
	ExpvarConnectsOK  int64
	ExpvarAnnounces   int64
	ExpvarAnnouncesOK int64
	ExpvarScrapes     int64
	ExpvarScrapesOK   int64
	ExpvarErrs        int64
	ExpvarClienterrs  int64
	ExpvarSeeds       int64
	ExpvarLeeches     int64
	ExpvarIPs         map[PeerIP]int8
)

func initExpvar() {
	// Might as well alloc capcity at start
	ExpvarIPs = make(map[PeerIP]int8, 30000)

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
