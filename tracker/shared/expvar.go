package shared

import "sync"

var (
	// These should only be accessed with atomic
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
	ExpvarIPs         = struct {
		sync.Mutex
		M map[PeerIP]int8
	}{M: make(map[PeerIP]int8, 30000)}
)

func initExpvar() {
	if ok := PeerDB.check(); !ok {
		panic("peerDB not init before expvars")
	}

	ExpvarIPs.Lock()
	for _, peermap := range PeerDB.db {
		for _, peer := range peermap {
			ExpvarIPs.M[peer.IP]++
			if peer.Complete == true {
				ExpvarSeeds++
			} else {
				ExpvarLeeches++
			}
		}
	}
	ExpvarIPs.Unlock()
}
