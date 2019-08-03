package shared

import "sync"

type expvarIPmap struct {
	sync.Mutex
	M map[PeerIP]int8
}

var (
	// These should only be accessed with atomic
	Expvar struct {
		Connects    int64
		ConnectsOK  int64
		Announces   int64
		AnnouncesOK int64
		Scrapes     int64
		ScrapesOK   int64
		Errs        int64
		Clienterrs  int64
		Seeds       int64
		Leeches     int64
		IPs         expvarIPmap
	}
)

// InitExpvar sets the expvar vars to the contents of the peer database
func InitExpvar(peerdb *PeerDatabase) {
	if ok := peerdb.check(); !ok {
		panic("peerDB not init before expvars")
	}

	Expvar.IPs.M = make(map[PeerIP]int8, 30000)

	Expvar.IPs.Lock()
	for _, peermap := range peerdb.db {
		for _, peer := range peermap {
			Expvar.IPs.M[peer.IP]++
			if peer.Complete == true {
				Expvar.Seeds++
			} else {
				Expvar.Leeches++
			}
		}
	}
	Expvar.IPs.Unlock()

	peerdb.generateMetrics()
}
