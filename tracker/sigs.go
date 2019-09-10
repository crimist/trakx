package tracker

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/syc0x00/trakx/tracker/database"
	"github.com/syc0x00/trakx/tracker/udp"

	"go.uber.org/zap"
)

var (
	SigStop = os.Interrupt
)

func handleSigs(peerdb database.Database, udptracker *udp.UDPTracker) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGUSR1)

	for {
		sig := <-c

		switch sig {
		case os.Interrupt, os.Kill, syscall.SIGTERM:
			logger.Info("Exiting")

			peerdb.Backup().SaveFull()
			if udptracker != nil {
				udptracker.WriteConns()
			}

			os.Exit(0)
		default:
			logger.Info("Got unknown sig", zap.Any("Signal", sig))
		}
		// os.Exit(128 + int(sig.(syscall.Signal)))
	}
}
