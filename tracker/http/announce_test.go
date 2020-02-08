package http

import (
	"math/rand"
	"net"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"

	_ "github.com/crimist/trakx/tracker/storage/map"
)

// go build -gcflags '-m' -o /dev/null ./... |& grep "moved to heap:"

func BenchmarkAnnounce200(b *testing.B) {
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
	tracker.conf.Tracker.Numwant.Limit = 200 // for peerlistpool

	db, err := storage.Open(tracker.conf)
	if err != nil {
		b.Error("failed to open storage", err)
		b.FailNow()
	}
	tracker.peerdb = db

	params := announceParams{
		compact:  true,
		nopeerid: true,
		noneleft: false,
		event:    "started",
		port:     "6969",
		hash:     "01234567890123456789",
		peerid:   "01234567890123456789",
		numwant:  "200",
	}
	ip := storage.PeerIP{1, 2, 3, 4}

	random20 := func() string {
		var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		b := make([]rune, 20)
		for i := range b {
			b[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		return string(b)
	}

	for i := 0; i < 200; i++ {
		params.peerid = random20()
		tracker.announce(conn, &params, ip)
	}

	gcp := debug.SetGCPercent(-1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.announce(conn, &params, ip)
	}
	b.StopTimer()

	runtime.GC()
	debug.SetGCPercent(gcp)
}
