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
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-signalChannel

		switch sig {
		case os.Interrupt, syscall.SIGTERM: // Exit
			config.Logger.Info("Received exit signal", zap.Any("signal", sig))

			udptracker.Shutdown()
			httptracker.Shutdown()

			if err := peerdb.Backup().Save(); err != nil {
				config.Logger.Error("Database save failed", zap.Error(err))
			}

			if err := udptracker.WriteConns(); err != nil {
				config.Logger.Error("UDP connections save failed", zap.Error(err))
			}

			os.Exit(exitSuccess)

		case syscall.SIGUSR1: // Save
			config.Logger.Info("Received save signal", zap.Any("signal", sig))

			if err := peerdb.Backup().Save(); err != nil {
				config.Logger.Error("Database save failed", zap.Error(err))
			}

			if err := udptracker.WriteConns(); err != nil {
				config.Logger.Error("UDP connections save failed", zap.Error(err))
			}

			config.Logger.Info("Saves successful")

		default:
			config.Logger.Info("Received unknown signal, ignoring", zap.Any("signal", sig))
		}
	}
}
