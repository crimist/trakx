package tracker

import (
	"expvar"
	"net/http"
	"time"

	"github.com/Syc0x00/Trakx/tracker/shared"
)

func getInfo() (peers, hashes, ips, seeds, leeches int64) {
	ipmap := make(map[string]bool)

	for _, peermap := range shared.PeerDB {
		peers += int64(len(peermap))

		for _, peer := range peermap {
			ipmap[peer.IP] = true
			if peer.Complete == true {
				seeds++
			} else {
				leeches++
			}
		}
	}

	hashes = int64(len(shared.PeerDB))
	ips = int64(len(ipmap))

	return
}

// Expvar is for netdata
func Expvar() {
	var announcesOld int64
	var scrapesOld int64
	var errsOld int64

	uniqueIP := expvar.NewInt("tracker.db.ips")
	uniqueHash := expvar.NewInt("tracker.db.hashes")
	uniquePeer := expvar.NewInt("tracker.db.peers")

	seeds := expvar.NewInt("tracker.db.seeds")
	leeches := expvar.NewInt("tracker.db.leeches")

	announces := expvar.NewInt("tracker.performance.announces")
	announcesSec := expvar.NewInt("tracker.performance.announces.sec")

	errors := expvar.NewInt("tracker.performance.errors")
	errorsSec := expvar.NewInt("tracker.performance.errors.sec")

	scrapes := expvar.NewInt("tracker.performance.scrapes")
	scrapesSec := expvar.NewInt("tracker.performance.scrapes.sec")

	go http.ListenAndServe("127.0.0.1:"+shared.ExpvarPort, nil) // only on localhost

	nextTime := time.Now().Truncate(time.Second)

	for {
		peers, hashes, ips, s, l := getInfo()
		uniqueIP.Set(ips)
		uniqueHash.Set(hashes)
		uniquePeer.Set(peers)

		seeds.Set(s)
		leeches.Set(l)

		announces.Set(shared.ExpvarAnnounces)
		announcesSec.Set(shared.ExpvarAnnounces - announcesOld)
		announcesOld = shared.ExpvarAnnounces

		errors.Set(shared.ExpvarErrs)
		errorsSec.Set(shared.ExpvarErrs - errsOld)
		errsOld = shared.ExpvarErrs

		scrapes.Set(shared.ExpvarScrapes)
		scrapesSec.Set(shared.ExpvarScrapes - scrapesOld)
		scrapesOld = shared.ExpvarScrapes

		nextTime = nextTime.Add(time.Second)
		time.Sleep(time.Until(nextTime))
	}
}
