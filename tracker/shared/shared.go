package shared

import (
	"time"

	"go.uber.org/zap"
)

var (
	PeerDB PeerDatabase
)

func Init(conf *Config, logger *zap.Logger) {
	PeerDB = PeerDatabase{
		conf:   conf,
		logger: logger,
	}

	PeerDB.Load()
	initExpvar()
	PeerDB.generateMetrics()

	// Start threads
	go RunOn(time.Duration(conf.Database.Peer.Write)*time.Second, PeerDB.WriteTmp)
	go RunOn(time.Duration(conf.Database.Peer.Trim)*time.Second, PeerDB.Trim)
	if conf.Tracker.MetricsInterval > 0 {
		go RunOn(time.Duration(conf.Tracker.MetricsInterval)*time.Second, PeerDB.generateMetrics)
	}
}
