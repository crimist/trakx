package http

import (
	"math/rand"
	"net"
	"runtime/debug"
	"testing"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"

	_ "github.com/crimist/trakx/tracker/storage/map"
)

// go build -gcflags '-m' -o /dev/null ./... |& grep "moved to heap:"

const (
	addr = ":12345"
)

func BenchmarkAnnounce(b *testing.B) {
	// open listen
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}

	// run accepter
	go func() {
		for {
			ln.Accept()
		}
	}()

	// establish connection
	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		panic(err)
	}

	// setup tracker
	tracker := HTTPTracker{}
	config.Conf.Database.Type = "gomap"
	config.Conf.Database.Backup = "none"
	config.Conf.Tracker.AnnounceFuzz = 1
	config.Conf.Tracker.Numwant.Limit = 300 // for peerlistpool

	// setup db
	db, err := storage.Open()
	if err != nil {
		b.Error("failed to open storage", err)
		b.FailNow()
	}
	tracker.peerdb = db

	// setup params
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

	// init entries in database
	for i := 0; i < 300; i++ {
		params.peerid = random20()
		tracker.announce(conn, &params, ip)
	}

	gcp := debug.SetGCPercent(-1)
	b.ResetTimer()

	// run benchmark
	for i := 0; i < b.N; i++ {
		params.peerid = random20()
		tracker.announce(conn, &params, ip)
	}

	// cleanup
	b.StopTimer()
	debug.SetGCPercent(gcp)
	ln.Close()
}
