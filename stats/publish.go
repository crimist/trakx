package stats

// import (
// 	"expvar"
// 	"runtime"
// 	"time"

// 	"github.com/crimist/trakx/config"
// 	"github.com/crimist/trakx/pools"
// 	"github.com/crimist/trakx/storage"
// 	"github.com/crimist/trakx/utils"
// 	"go.uber.org/zap"
// )

// var initTime = time.Now()

// // Publish starts publishing and updating expvar values, requests metrics are over duration of Config.ExpvarInterval
// func Publish(peerdb storage.Database, udpconns func() int64) {
// 	zap.L().Info("publishing stats as expvars", zap.Duration("interval", config.Config.ExpvarInterval))

// 	// requests
// 	hits := expvar.NewInt("trakx.requests.hits")
// 	connects := expvar.NewInt("trakx.requests.connects")
// 	announces := expvar.NewInt("trakx.requests.announces")
// 	scrapes := expvar.NewInt("trakx.requests.scrapes")

// 	// database
// 	seeds := expvar.NewInt("trakx.database.seeds")
// 	leeches := expvar.NewInt("trakx.database.leeches")
// 	peers := expvar.NewInt("trakx.database.peers")
// 	ips := expvar.NewInt("trakx.database.ips")
// 	hashes := expvar.NewInt("trakx.database.hashes")
// 	udpConnections := expvar.NewInt("trakx.database.udpconnections")

// 	// errors
// 	serverErrors := expvar.NewInt("trakx.errors.server")
// 	clientErrors := expvar.NewInt("trakx.errors.client")

// 	// pools
// 	dictionaryPool := expvar.NewInt("trakx.pools.dictionaries")
// 	peerPool := expvar.NewInt("trakx.pools.peers")
// 	peerlist4Pool := expvar.NewInt("trakx.pools.peerlists4")
// 	peerlist6Pool := expvar.NewInt("trakx.pools.peerlists6")

// 	// internal
// 	goroutines := expvar.NewInt("trakx.internal.goroutines")
// 	uptime := expvar.NewInt("trakx.internal.uptime")

// 	utils.RunOn(config.Config.ExpvarInterval, func() {
// 		// set expvars
// 		hits.Set(Hits.Load())
// 		connects.Set(Connects.Load())
// 		announces.Set(Announces.Load())
// 		scrapes.Set(Scrapes.Load())

// 		seeds.Set(Seeds.Load())
// 		leeches.Set(Leeches.Load())
// 		peers.Set(Seeds.Load() + Leeches.Load())
// 		ips.Set(int64(IPStats.Total()))
// 		hashes.Set(int64(peerdb.Hashes()))
// 		udpConnections.Set(udpconns())

// 		serverErrors.Set(ServerErrors.Load())
// 		clientErrors.Set(ClientErrors.Load())

// 		dictionaryPool.Set(int64(pools.Dictionaries.Created()))
// 		peerPool.Set(int64(pools.Peers.Created()))
// 		peerlist4Pool.Set(int64(pools.Peerlists4.Created()))
// 		peerlist6Pool.Set(int64(pools.Peerlists6.Created()))

// 		goroutines.Set(int64(runtime.NumGoroutine()))
// 		uptime.Set(int64(time.Since(initTime) / time.Second))

// 		// reset requests / s stats
// 		Hits.Store(0)
// 		Connects.Store(0)
// 		Announces.Store(0)
// 		Scrapes.Store(0)
// 	})
// }
