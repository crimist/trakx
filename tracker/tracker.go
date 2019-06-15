package tracker

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	trackerCleanTimeout     = 45 * time.Minute
	trackerAnnounceInterval = 20 * time.Minute
	trackerCleanInterval    = 3 * time.Minute
	trackerWriteDBInterval  = 5 * time.Minute
	trackerDBFilename       = "trakx.db"
)

var (
	db     Database
	logger *zap.Logger
	env    Enviroment
)

// Init initiates all the things the tracker needs
func Init(isProd bool) error {
	var err error
	var cfg zap.Config
	db = make(Database)

	if isProd == true {
		env = Prod
		cfg = zap.NewProductionConfig()
	} else {
		env = Dev
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.OutputPaths = append(cfg.OutputPaths, "trakx.log")
	logger, err = cfg.Build()

	db.Load()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-c
		logger.Info("Got signal", zap.Any("Signal", sig))

		db.Write()

		os.Exit(128 + int(sig.(syscall.Signal)))
	}()

	go func() {
		for c := time.Tick(trackerWriteDBInterval); ; <-c {
			db.Write()
		}
	}()

	return err
}

// Cleaner removes clients that haven't checked in recently
func Cleaner() {
	for c := time.Tick(trackerCleanInterval); ; <-c {
		db.Clean()
	}
}
