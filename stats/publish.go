package stats

import (
	"expvar"
	"runtime"
	"time"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/utils"
	"go.uber.org/zap"
)

var initTime = time.Now()

type PeriodicConfig struct {
	Collector               Collector
	Interval                time.Duration
	GetHashes               func() int
	GetConns                func() int
	GetPeerlist4PoolCreated func() int64
	GetPeerlist6PoolCreated func() int64
}

// PublishPeriodic starts a goroutine to publish expensive-to-calculate stats at a given interval.
// It also publishes some general stats like pool sizes and uptime.
func PublishPeriodic(cfg PeriodicConfig) {
	zap.L().Info("publishing statistics", zap.Duration("interval", cfg.Interval))

	// database
	peers := expvar.NewInt("trakx.database.peers")
	ips := expvar.NewInt("trakx.database.ips")
	hashes := expvar.NewInt("trakx.database.hashes")
	connections := expvar.NewInt("trakx.database.connections")

	// pools
	dictionaryPool := expvar.NewInt("trakx.pools.dictionaries")
	peerlist4Pool := expvar.NewInt("trakx.pools.peerlists4")
	peerlist6Pool := expvar.NewInt("trakx.pools.peerlists6")

	// internal
	goroutines := expvar.NewInt("trakx.internal.goroutines")
	uptime := expvar.NewInt("trakx.internal.uptime")

	utils.RunOn(cfg.Interval, func() {
		snapshot := cfg.Collector.GetSnapshot()

		peers.Set(snapshot.Seeds + snapshot.Leeches)
		ips.Set(int64(snapshot.IPs))
		hashes.Set(int64(cfg.GetHashes()))
		if cfg.GetConns != nil {
			connections.Set(int64(cfg.GetConns()))
		}

		dictionaryPool.Set(bencoding.DictionaryPoolCreated())
		if cfg.GetPeerlist4PoolCreated != nil {
			peerlist4Pool.Set(cfg.GetPeerlist4PoolCreated())
		} else {
			peerlist4Pool.Set(0)
		}
		if cfg.GetPeerlist6PoolCreated != nil {
			peerlist6Pool.Set(cfg.GetPeerlist6PoolCreated())
		} else {
			peerlist6Pool.Set(0)
		}

		goroutines.Set(int64(runtime.NumGoroutine()))
		uptime.Set(int64(time.Since(initTime) / time.Second))
	})
}
