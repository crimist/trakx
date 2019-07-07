package shared

import (
	"go.uber.org/zap"
)

func Init(prod bool) error {
	UDPCheckConnID = true
	setEnv(prod)
	if err := setLogger(prod); err != nil {
		return err
	}
	PeerDB.Load()
	initExpvar()

	go RunOn(WriteDBInterval, PeerDB.WriteTmp)
	go RunOn(CleanInterval, PeerDB.Clean)

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
