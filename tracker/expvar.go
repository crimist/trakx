package tracker

import (
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
)

func publishExpvar(conf *shared.Config) {
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
	go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", conf.Trakx.Expvar.Port), nil)

	shared.RunOn(time.Duration(conf.Trakx.Expvar.Every)*time.Second, func() {
		shared.Expvar.IPs.Lock()
		uniqueIP.Set(int64(len(shared.Expvar.IPs.M)))
		shared.Expvar.IPs.Unlock()
		uniqueHash.Set(int64(shared.PeerDB.Hashes()))

		s := atomic.LoadInt64(&shared.Expvar.Seeds)
		l := atomic.LoadInt64(&shared.Expvar.Leeches)
		uniquePeer.Set(s + l)
		seeds.Set(s)
		leeches.Set(l)

		// database
		conns.Set(int64(udptracker.GetConnCount()))

		announcesSec.Set(atomic.LoadInt64(&shared.Expvar.Announces))
		atomic.StoreInt64(&shared.Expvar.Announces, 0)

		announcesSecOK.Set(atomic.LoadInt64(&shared.Expvar.AnnouncesOK))
		atomic.StoreInt64(&shared.Expvar.AnnouncesOK, 0)

		scrapesSec.Set(atomic.LoadInt64(&shared.Expvar.Scrapes))
		atomic.StoreInt64(&shared.Expvar.Scrapes, 0)

		scrapesSecOK.Set(atomic.LoadInt64(&shared.Expvar.ScrapesOK))
		atomic.StoreInt64(&shared.Expvar.ScrapesOK, 0)

		clientErrors.Set(atomic.LoadInt64(&shared.Expvar.Clienterrs))
		atomic.StoreInt64(&shared.Expvar.Clienterrs, 0)

		e := atomic.LoadInt64(&shared.Expvar.Errs)
		errors.Set(e)
		errorsSec.Set(e - errsOld)
		errsOld = e

		connects.Set(atomic.LoadInt64(&shared.Expvar.Connects))
		atomic.StoreInt64(&shared.Expvar.Connects, 0)

		connectsOK.Set(atomic.LoadInt64(&shared.Expvar.ConnectsOK))
		atomic.StoreInt64(&shared.Expvar.ConnectsOK, 0)
	})
}
