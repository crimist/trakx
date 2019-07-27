package tracker

import (
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
)

// Expvar is for netdata
func Expvar() {
	var errsOld int64

	// Stats
	uniqueIP := expvar.NewInt("tracker.stats.ips")
	uniqueHash := expvar.NewInt("tracker.stats.hashes")
	uniquePeer := expvar.NewInt("tracker.stats.peers")
	seeds := expvar.NewInt("tracker.stats.seeds")
	leeches := expvar.NewInt("tracker.stats.leeches")

	// Performance
	announcesSec := expvar.NewInt("tracker.performance.announces")
	announcesSecOK := expvar.NewInt("tracker.performance.announcesok")
	errors := expvar.NewInt("tracker.performance.errors")
	errorsSec := expvar.NewInt("tracker.performance.errorssec")
	clineterrs := expvar.NewInt("tracker.performance.clienterrs")
	scrapesSec := expvar.NewInt("tracker.performance.scrapes")
	scrapesSecOK := expvar.NewInt("tracker.performance.scrapesok")
	connects := expvar.NewInt("tracker.performance.connects")
	connectsOK := expvar.NewInt("tracker.performance.connectsok")

	// only listen on localhost
	go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", shared.Config.Tracker.Ports.Expvar), nil)

	nextTime := time.Now().Truncate(time.Second)

	for {
		uniqueIP.Set(int64(len(shared.ExpvarIPs)))
		uniqueHash.Set(int64(len(shared.PeerDB)))
		uniquePeer.Set(shared.ExpvarSeeds + shared.ExpvarLeeches)
		seeds.Set(shared.ExpvarSeeds)
		leeches.Set(shared.ExpvarLeeches)
		announcesSec.Set(shared.ExpvarAnnounces)
		shared.ExpvarAnnounces = 0
		announcesSecOK.Set(shared.ExpvarAnnouncesOK)
		shared.ExpvarAnnouncesOK = 0
		scrapesSec.Set(shared.ExpvarScrapes)
		shared.ExpvarScrapes = 0
		scrapesSecOK.Set(shared.ExpvarScrapesOK)
		shared.ExpvarScrapesOK = 0
		clineterrs.Set(shared.ExpvarClienterrs)
		shared.ExpvarClienterrs = 0
		errors.Set(shared.ExpvarErrs)
		errorsSec.Set(shared.ExpvarErrs - errsOld)
		errsOld = shared.ExpvarErrs
		connects.Set(shared.ExpvarConnects)
		shared.ExpvarConnects = 0
		connectsOK.Set(shared.ExpvarConnectsOK)
		shared.ExpvarConnectsOK = 0

		nextTime = nextTime.Add(time.Second)
		time.Sleep(time.Until(nextTime))
	}
}
