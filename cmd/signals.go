package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker"

	"go.uber.org/zap"
)

// SigStop is the signal which Trakx uses to shutdwn gracefully
var SigStop = os.Interrupt

const exitSuccess = 0

func signalHandler(db storage.Database, trackers []tracker.Tracker) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-signalChannel

		switch sig {
		case os.Interrupt, syscall.SIGTERM: // Exit
			zap.L().Info("Received exit signal", zap.Any("signal", sig))

			for _, tracker := range trackers {
				tracker.Shutdown()
			}

			// TODO: write db
			// TODO: write udp conn db

			os.Exit(exitSuccess)

		case syscall.SIGUSR1: // persist
			zap.L().Info("Received persist signal", zap.Any("signal", sig))

			// TODO: write db
			// TODO: write udp conn db

			zap.L().Info("Persisted databases")

		default:
			zap.L().Info("Received unknown signal, ignoring", zap.Any("signal", sig))
		}
	}
}
