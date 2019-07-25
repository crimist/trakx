package tracker

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Syc0x00/Trakx/tracker/shared"
	udptracker "github.com/Syc0x00/Trakx/tracker/udp"
	"go.uber.org/zap"
)

func handleSigs() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-c

		switch sig {
		case os.Interrupt, os.Kill:
			shared.Logger.Info("Exiting")

			shared.PeerDB.WriteFull()
			udptracker.WriteConnDB()
			os.Exit(128 + int(sig.(syscall.Signal)))
		case syscall.SIGUSR1:
			shared.Logger.Info("Reloading config (not all setting will apply until restart)")
			shared.LoadConfig()
		default:
			shared.Logger.Info("Got unknown sig", zap.Any("Signal", sig))
		}
	}
}
