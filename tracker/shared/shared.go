package shared

import (
	"time"

	"go.uber.org/zap"
)

func Init() error {
	UDPCheckConnID = true // TODO move to config

	loadConfig()
	setEnv(Config.Production)
	if err := setLogger(Config.Production); err != nil {
		return err
	}
	PeerDB.Load()
	initExpvar()
	processMetrics()

	// Start threads
	go RunOn(time.Duration(Config.DBWriteInterval)*time.Second, PeerDB.WriteTmp)
	go RunOn(time.Duration(Config.DBCleanInterval)*time.Second, PeerDB.Clean)
	if Config.MetricsInterval > 0 {
		go RunOn(time.Duration(Config.MetricsInterval)*time.Second, processMetrics)
	}

	return nil
}

func setEnv(prod bool) {
	if prod == true {
		Env = Prod
	} else {
		Env = Dev
	}
}

func setLogger(prod bool) error {
	var err error
	var cfg zap.Config

	if prod == true {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.OutputPaths = append(cfg.OutputPaths, "trakx.log")
	}
	Logger, err = cfg.Build()
	return err
}
