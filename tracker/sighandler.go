package tracker

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp"

	"go.uber.org/zap"
)

// SigStop is the signal which Trakx uses to shutdwn gracefully
var SigStop = os.Interrupt

const exitSuccess = 0

func signalHandler(peerdb storage.Database, udptracker *udp.UDPTracker, httptracker *http.HTTPTracker) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-c

		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			// Exit
			config.Logger.Info("Got exit signal", zap.Any("sig", sig))

			udptracker.Shutdown()
			httptracker.Shutdown()

			if err := peerdb.Backup().Save(); err != nil {
				config.Logger.Info("Failed to backup the database on exit")
			}

			udptracker.WriteConns()

			config.Logger.Info("Exiting")
			os.Exit(exitSuccess)
		case syscall.SIGUSR1:
			// Save
			config.Logger.Info("Got save signal", zap.Any("sig", sig))

			if err := peerdb.Backup().Save(); err != nil {
				config.Logger.Info("Failed to backup the database on save")
			}

			udptracker.WriteConns()

			config.Logger.Info("Saved")
		default:
			config.Logger.Info("Got unknown sig", zap.Any("sig", sig))
		}
		// os.Exit(128 + int(sig.(syscall.Signal)))
	}
}
