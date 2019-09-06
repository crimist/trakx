package tracker

import (
	"expvar"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/syc0x00/trakx/tracker/database"
	"github.com/syc0x00/trakx/tracker/http"
	"github.com/syc0x00/trakx/tracker/shared"
	"github.com/syc0x00/trakx/tracker/udp"
)

func publishExpvar(conf *shared.Config, peerdb database.Database, httptracker *http.HTTPTracker, udptracker *udp.UDPTracker) {
	var errsOld int64
	start := time.Now()

	// Stats
	uniqueIP := expvar.NewInt("tracker.stats.ips")
	uniqueHash := expvar.NewInt("tracker.stats.hashes")
	uniquePeer := expvar.NewInt("tracker.stats.peers")
	seeds := expvar.NewInt("tracker.stats.seeds")
	leeches := expvar.NewInt("tracker.stats.leeches")

	// database
	conns := expvar.NewInt("trakx.database.connections")
	uptime := expvar.NewInt("trakx.database.uptime")

	// Performance
	goroutines := expvar.NewInt("trakx.performance.goroutines")
	qlen := expvar.NewInt("trakx.performance.qlen")

	announcesSec := expvar.NewInt("tracker.performance.announces")
	announcesSecOK := expvar.NewInt("tracker.performance.announcesok")
	errors := expvar.NewInt("tracker.performance.errors")
	errorsSec := expvar.NewInt("tracker.performance.errorssec")
	clientErrors := expvar.NewInt("tracker.performance.clienterrs")
	scrapesSec := expvar.NewInt("tracker.performance.scrapes")
	scrapesSecOK := expvar.NewInt("tracker.performance.scrapesok")
	connects := expvar.NewInt("tracker.performance.connects")
	connectsOK := expvar.NewInt("tracker.performance.connectsok")

	shared.RunOn(time.Duration(conf.Trakx.Expvar.Every)*time.Second, func() {
		database.Expvar.IPs.Lock()
		uniqueIP.Set(int64(len(database.Expvar.IPs.M)))
		database.Expvar.IPs.Unlock()
		uniqueHash.Set(int64(peerdb.Hashes()))

		s := atomic.LoadInt64(&database.Expvar.Seeds)
		l := atomic.LoadInt64(&database.Expvar.Leeches)
		uniquePeer.Set(s + l)
		seeds.Set(s)
		leeches.Set(l)

		// database
		if udptracker != nil {
			conns.Set(int64(udptracker.GetConnCount()))
		}
		uptime.Set(int64(time.Since(start) / time.Second))

		// performance
		goroutines.Set(int64(runtime.NumGoroutine()))
		qlen.Set(int64(httptracker.QueueLen()))

		announcesSec.Set(atomic.LoadInt64(&database.Expvar.Announces))
		atomic.StoreInt64(&database.Expvar.Announces, 0)

		announcesSecOK.Set(atomic.LoadInt64(&database.Expvar.AnnouncesOK))
		atomic.StoreInt64(&database.Expvar.AnnouncesOK, 0)

		scrapesSec.Set(atomic.LoadInt64(&database.Expvar.Scrapes))
		atomic.StoreInt64(&database.Expvar.Scrapes, 0)

		scrapesSecOK.Set(atomic.LoadInt64(&database.Expvar.ScrapesOK))
		atomic.StoreInt64(&database.Expvar.ScrapesOK, 0)

		clientErrors.Set(atomic.LoadInt64(&database.Expvar.Clienterrs))
		atomic.StoreInt64(&database.Expvar.Clienterrs, 0)

		e := atomic.LoadInt64(&database.Expvar.Errs)
		errors.Set(e)
		errorsSec.Set(e - errsOld)
		errsOld = e

		connects.Set(atomic.LoadInt64(&database.Expvar.Connects))
		atomic.StoreInt64(&database.Expvar.Connects, 0)

		connectsOK.Set(atomic.LoadInt64(&database.Expvar.ConnectsOK))
		atomic.StoreInt64(&database.Expvar.ConnectsOK, 0)
	})
}
