package shared

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

var (
	PeerDB PeerDatabase
	Logger *zap.Logger
	Env    Enviroment
)

func Init(prod bool) error {
	setEnv(prod)
	if err := setLogger(prod); err != nil {
		return err
	}
	PeerDB.Load()
	setSignals()
	initExpvar()

	go RunOn(WriteDBInterval, PeerDB.WriteTmp)
	go RunOn(CleanInterval, PeerDB.Clean)

	return nil
}

func setSignals() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-c
		Logger.Info("Got signal", zap.Any("Signal", sig))

		PeerDB.WriteFull()

		os.Exit(128 + int(sig.(syscall.Signal)))
	}()
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
