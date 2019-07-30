package shared

import (
	"time"

	"go.uber.org/zap"
)

func Init() error {
	LoadConfig()

	// get logger based on config
	if err := setLogger(Config.Trakx.Prod); err != nil {
		return err
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

func setLogger(prod bool) error {
	var err error
	var cfg zap.Config

	if prod == true {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}
	Logger, err = cfg.Build()
	return err
}
