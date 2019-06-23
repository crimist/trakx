package tracker

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	trackerExpvarPort      = "1338"
	trackerCleanTimeout    = 45 * 60 // 45 min
	AnnounceInterval       = 20 * 60 // 20 min
	trackerCleanInterval   = 3 * time.Minute
	trackerWriteDBInterval = 5 * time.Minute
	trackerDBFilename      = "trakx.db"
	trackerDBTempFilename  = "trakx.db.tmp"
	DefaultNumwant         = 300
)

var (
	PeerDB Database
	Logger *zap.Logger
	Env    Enviroment
)

// Init initiates all the things the tracker needs
func Init(isProd bool) error {
	var err error
	var cfg zap.Config
	PeerDB = make(Database)

	if isProd == true {
		Env = Prod
		cfg = zap.NewProductionConfig()
	} else {
		Env = Dev
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.OutputPaths = append(cfg.OutputPaths, "trakx.log")
	Logger, err = cfg.Build()
	if err != nil {
		return err
	}

	PeerDB.Load()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-c
		Logger.Info("Got signal", zap.Any("Signal", sig))

		PeerDB.Write(false)

		os.Exit(128 + int(sig.(syscall.Signal)))
	}()

	go Writer()

	return err
}

// Writer runs db.Write() every trackerWriteDBInterval
func Writer() {
	time.Sleep(1 * time.Second)
	for c := time.Tick(trackerWriteDBInterval); ; <-c {
		PeerDB.Write(true)
	}
}

// Cleaner removes clients that haven't checked in recently
func Cleaner() {
	time.Sleep(1 * time.Second)
	for c := time.Tick(trackerCleanInterval); ; <-c {
		PeerDB.Clean()
	}
}
