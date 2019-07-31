package tracker

import (
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
	"github.com/syc0x00/trakx/tracker/udp"
)

func publishExpvar() {
	if shared.Config.Trakx.Expvar.Every < 1 {
		shared.Logger.Panic("Expvar.Every < 1")
	}

	var errsOld int64

	// Stats
	uniqueIP := expvar.NewInt("tracker.stats.ips")
	uniqueHash := expvar.NewInt("tracker.stats.hashes")
	uniquePeer := expvar.NewInt("tracker.stats.peers")
	seeds := expvar.NewInt("tracker.stats.seeds")
	leeches := expvar.NewInt("tracker.stats.leeches")

	// database
	conns := expvar.NewInt("trakx.database.connections")

	// Performance
	announcesSec := expvar.NewInt("tracker.performance.announces")
	announcesSecOK := expvar.NewInt("tracker.performance.announcesok")
	errors := expvar.NewInt("tracker.performance.errors")
	errorsSec := expvar.NewInt("tracker.performance.errorssec")
	clientErrors := expvar.NewInt("tracker.performance.clienterrs")
	scrapesSec := expvar.NewInt("tracker.performance.scrapes")
	scrapesSecOK := expvar.NewInt("tracker.performance.scrapesok")
	connects := expvar.NewInt("tracker.performance.connects")
	connectsOK := expvar.NewInt("tracker.performance.connectsok")

	// only listen on localhost
	go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", shared.Config.Trakx.Expvar.Port), nil)

	shared.RunOn(time.Duration(shared.Config.Trakx.Expvar.Every)*time.Second, func() {
		shared.ExpvarIPs.Lock()
		uniqueIP.Set(int64(len(shared.ExpvarIPs.M)))
		shared.ExpvarIPs.Unlock()
		uniqueHash.Set(int64(shared.PeerDB.Hashes()))

		s := atomic.LoadInt64(&shared.ExpvarSeeds)
		l := atomic.LoadInt64(&shared.ExpvarLeeches)
		uniquePeer.Set(s + l)
		seeds.Set(s)
		leeches.Set(l)

		// database
		conns.Set(int64(udp.GetConnCount()))

		announcesSec.Set(atomic.LoadInt64(&shared.ExpvarAnnounces))
		atomic.StoreInt64(&shared.ExpvarAnnounces, 0)

		announcesSecOK.Set(atomic.LoadInt64(&shared.ExpvarAnnouncesOK))
		atomic.StoreInt64(&shared.ExpvarAnnouncesOK, 0)

		scrapesSec.Set(atomic.LoadInt64(&shared.ExpvarScrapes))
		atomic.StoreInt64(&shared.ExpvarScrapes, 0)

		scrapesSecOK.Set(atomic.LoadInt64(&shared.ExpvarScrapesOK))
		atomic.StoreInt64(&shared.ExpvarScrapesOK, 0)

		clientErrors.Set(atomic.LoadInt64(&shared.ExpvarClienterrs))
		atomic.StoreInt64(&shared.ExpvarClienterrs, 0)

		e := atomic.LoadInt64(&shared.ExpvarErrs)
		errors.Set(e)
		errorsSec.Set(e - errsOld)
		errsOld = e

		connects.Set(atomic.LoadInt64(&shared.ExpvarConnects))
		atomic.StoreInt64(&shared.ExpvarConnects, 0)

		connectsOK.Set(atomic.LoadInt64(&shared.ExpvarConnectsOK))
		atomic.StoreInt64(&shared.ExpvarConnectsOK, 0)
	})
}
