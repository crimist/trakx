package http

import (
	"net"
	"testing"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"

	_ "github.com/crimist/trakx/tracker/storage/map"
)

// go build -gcflags '-m' -o /dev/null ./... |& grep "moved to heap:"

func BenchmarkAnnounce(b *testing.B) {
	conn, _ := net.Dial("udp", ":1")

	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{} // silence
	logger, err := cfg.Build()
	if err != nil {
		b.Error("failed to build zap", err)
		b.FailNow()
	}

	tracker := HTTPTracker{}
	tracker.conf = &shared.Config{}
	tracker.conf.Logger = logger
	tracker.conf.Database.Type = "gomap"
	tracker.conf.Database.Backup = "file"
	tracker.conf.Tracker.AnnounceFuzz = 1

	db, err := storage.Open(tracker.conf)
	if err != nil {
		b.Error("failed to open storage", err)
		b.FailNow()
	}
	tracker.peerdb = db

	ip := storage.PeerIP{1, 2, 3, 4}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params := announceParams{
			compact:  true,
			nopeerid: true,
			noneleft: false,
			event:    "started",
			port:     "6969",
			hash:     "01234567890123456789",
			peerid:   "01234567890123456789",
			numwant:  "20",
		}
		tracker.announce(conn, &params, ip)
	}
}
