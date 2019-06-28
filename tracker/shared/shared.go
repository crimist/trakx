package shared

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	// httptracker "github.com/Syc0x00/Trakx/tracker/http"
)

const (
	HTTPPort           = "1337"
	UDPPort            = 1337
	ExpvarPort         = "1338"
	AnnounceInterval   = 30 * 60 // 30 min
	CleanTimeout       = AnnounceInterval * 2
	CleanInterval      = 3 * time.Minute
	WriteDBInterval    = 5 * time.Minute
	PeerDBFilename     = "trakx.db"
	PeerDBTempFilename = "trakx.db.tmp"
	DefaultNumwant     = 300
	Bye                = "See you space cowboy..."
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

	go Writer()
	go Cleaner()

	return nil
}

func setSignals() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-c
		Logger.Info("Got signal", zap.Any("Signal", sig))

		PeerDB.Write(false)

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
	}
	cfg.OutputPaths = append(cfg.OutputPaths, "trakx.log")
	Logger, err = cfg.Build()
	return err
}

// Writer runs db.Write() every WriteDBInterval
func Writer() {
	time.Sleep(1 * time.Second)
	for c := time.Tick(WriteDBInterval); ; <-c {
		PeerDB.Write(true)
	}
}

// Cleaner removes clients that haven't checked in recently
func Cleaner() {
	time.Sleep(1 * time.Second)
	for c := time.Tick(CleanInterval); ; <-c {
		PeerDB.Clean()
	}
}
