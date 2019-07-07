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

		if sig == os.Interrupt || sig == os.Kill || sig == syscall.SIGTERM {
			shared.Logger.Info("Exiting", zap.Any("Signal", sig))

			shared.PeerDB.WriteFull()
			udptracker.WriteConnDB()
			os.Exit(128 + int(sig.(syscall.Signal)))
		} else if sig == syscall.SIGUSR1 {
			shared.Logger.Info("Toggling connID check", zap.Any("Signal", sig))

			shared.UDPCheckConnID = !shared.UDPCheckConnID
		} else {
			shared.Logger.Info("Got unknown sig", zap.Any("Signal", sig))
		}
	}
}
