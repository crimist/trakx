package tracker

import (
	"expvar"
	"net/http"
	"time"

	"github.com/Syc0x00/Trakx/tracker/shared"
)

// Expvar is for netdata
func Expvar() {
	var errsOld int64

	uniqueIP := expvar.NewInt("tracker.db.ips")
	uniqueHash := expvar.NewInt("tracker.db.hashes")
	uniquePeer := expvar.NewInt("tracker.db.peers")
	seeds := expvar.NewInt("tracker.db.seeds")
	leeches := expvar.NewInt("tracker.db.leeches")
	announcesSec := expvar.NewInt("tracker.performance.announces.sec")
	announcesSecOK := expvar.NewInt("tracker.performance.announces.sec.ok")
	errors := expvar.NewInt("tracker.performance.errors")
	errorsSec := expvar.NewInt("tracker.performance.errors.sec")
	clineterrs := expvar.NewInt("tracker.performance.clienterrs")
	scrapesSec := expvar.NewInt("tracker.performance.scrapes.sec")
	scrapesSecOK := expvar.NewInt("tracker.performance.scrapes.sec.ok")

	go http.ListenAndServe("127.0.0.1:"+shared.ExpvarPort, nil) // only on localhost

	nextTime := time.Now().Truncate(time.Second)

	for {
		uniqueIP.Set(int64(len(shared.ExpvarIPs)))
		uniqueHash.Set(int64(len(shared.PeerDB)))
		uniquePeer.Set(shared.ExpvarSeeds + shared.ExpvarLeeches)
		seeds.Set(shared.ExpvarSeeds)
		leeches.Set(shared.ExpvarLeeches)
		announcesSec.Set(shared.ExpvarAnnounces)
		announcesSecOK.Set(shared.ExpvarAnnouncesOK)
		scrapesSec.Set(shared.ExpvarScrapes)
		scrapesSecOK.Set(shared.ExpvarScrapesOK)
		clineterrs.Set(shared.ExpvarClienterrs)
		errors.Set(shared.ExpvarErrs)
		errorsSec.Set(shared.ExpvarErrs - errsOld)
		errsOld = shared.ExpvarErrs

		shared.ExpvarAnnounces = 0
		shared.ExpvarAnnouncesOK = 0
		shared.ExpvarClienterrs = 0
		shared.ExpvarScrapes = 0
		shared.ExpvarScrapesOK = 0

		nextTime = nextTime.Add(time.Second)
		time.Sleep(time.Until(nextTime))
	}
}
