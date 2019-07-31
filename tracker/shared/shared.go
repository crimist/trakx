package shared

import (
	"time"

	"go.uber.org/zap"
)

func Init() error {
	var cfg zap.Config
	var err error

	LoadConfig()

	if Config.Trakx.Prod {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	if Logger, err = cfg.Build(); err != nil {
		panic(err)
	}

	PeerDB.Load()
	initExpvar()
	processMetrics()

	// Start threads
	go RunOn(time.Duration(Config.Database.Peer.Write)*time.Second, PeerDB.WriteTmp)
	go RunOn(time.Duration(Config.Database.Peer.Trim)*time.Second, PeerDB.Trim)
	if Config.Tracker.MetricsInterval > 0 {
		go RunOn(time.Duration(Config.Tracker.MetricsInterval)*time.Second, processMetrics)
	}

	return nil
}
