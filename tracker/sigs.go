package tracker

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp"
	"github.com/honeybadger-io/honeybadger-go"

	"go.uber.org/zap"
)

var (
	SigStop     = os.Interrupt
	exitSuccess = 0
)

func sigHandler(peerdb storage.Database, udptracker *udp.UDPTracker) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-c

		switch sig {
		case os.Interrupt, os.Kill, syscall.SIGTERM:
			// Exit
			logger.Info("Got exit signal", zap.Any("sig", sig))

			if err := peerdb.Backup().Save(); err != nil {
				honeybadger.Notify(err)
			}

			if udptracker != nil {
				udptracker.WriteConns()
			}

			os.Exit(exitSuccess)
		case syscall.SIGUSR1:
			// Save
			logger.Info("Got save signal", zap.Any("sig", sig))

			if err := peerdb.Backup().Save(); err != nil {
				honeybadger.Notify(err)
			}

			if udptracker != nil {
				udptracker.WriteConns()
			}
		default:
			logger.Info("Got unknown sig", zap.Any("sig", sig))
		}
		// os.Exit(128 + int(sig.(syscall.Signal)))
	}
}
