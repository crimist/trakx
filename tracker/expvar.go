package tracker

import (
	"expvar"
	"runtime"
	"time"

	"github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp"
)

var start = time.Now()

// TODO: Adhere to `fast` tag
func publishExpvar(conf *shared.Config, peerdb storage.Database, httptracker *http.HTTPTracker, udptracker *udp.UDPTracker) {
	// database
	ips := expvar.NewInt("trakx.database.ips")
	hashes := expvar.NewInt("trakx.database.hashes")
	peers := expvar.NewInt("trakx.database.peers")

	// stats
	connections := expvar.NewInt("trakx.stats.udpconnections")
	goroutines := expvar.NewInt("trakx.stats.goroutines")
	uptime := expvar.NewInt("trakx.stats.uptime")

	shared.RunOn(time.Duration(conf.Trakx.Expvar.Every)*time.Second, func() {
		storage.Expvar.IPs.Lock()
		ips.Set(int64(storage.Expvar.IPs.Len()))
		storage.Expvar.IPs.Unlock()
		hashes.Set(int64(peerdb.Hashes()))
		peers.Set(storage.Expvar.Seeds.Value() + storage.Expvar.Leeches.Value())

		// stats
		if udptracker != nil {
			connections.Set(int64(udptracker.ConnCount()))
		}
		uptime.Set(int64(time.Since(start) / time.Second))
		goroutines.Set(int64(runtime.NumGoroutine()))

		// reset per second values
		storage.Expvar.Hits.Set(0)
		storage.Expvar.Announces.Set(0)
		storage.Expvar.AnnouncesOK.Set(0)
		storage.Expvar.Scrapes.Set(0)
		storage.Expvar.ScrapesOK.Set(0)
		storage.Expvar.ClientErrors.Set(0)
		storage.Expvar.Connects.Set(0)
		storage.Expvar.ConnectsOK.Set(0)
	})
}
