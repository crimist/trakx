package shared

import (
	"time"

	"go.uber.org/zap"
)

func Init() error {
	var err error

	LoadConfig()

	cfg := zap.NewDevelopmentConfig()
	if Logger, err = cfg.Build(); err != nil {
		panic(err)
	}

	PeerDB.Load()
	initExpvar()
	PeerDB.generateMetrics()

	// Start threads
	go RunOn(time.Duration(Config.Database.Peer.Write)*time.Second, PeerDB.WriteTmp)
	go RunOn(time.Duration(Config.Database.Peer.Trim)*time.Second, PeerDB.Trim)
	if Config.Tracker.MetricsInterval > 0 {
		go RunOn(time.Duration(Config.Tracker.MetricsInterval)*time.Second, PeerDB.generateMetrics)
	}

	return nil
}
