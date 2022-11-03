package http

import (
	"math/rand"
	"net/netip"
	"testing"
	"time"

	"github.com/cbeuw/connutil"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
)

func randString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func BenchmarkHTTPAnnounceCompact200(b *testing.B) {
	// setup pipe
	client, server := connutil.AsyncPipe()
	defer func() {
		client.Close()
		server.Close()
	}()

	// config
	tracker := HTTPTracker{}
	config.Config.DB.Type = "gomap"
	config.Config.DB.Backup.Type = "none"
	config.Config.Announce.Fuzz = 1 * time.Second
	config.Config.Numwant.Limit = 300

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
	addr := netip.MustParseAddr("123.123.123.123")

	// init entries in database
	for i := 0; i < 300; i++ {
		params.peerid = randString(20)
		tracker.announce(client, &params, addr)
	}

	b.ResetTimer()

	// run benchmark
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		params.peerid = randString(20)
		b.StartTimer()
		tracker.announce(client, &params, addr)
	}
}

func BenchmarkHTTPAnnounce200(b *testing.B) {
	// open listen
	// setup pipe
	client, server := connutil.AsyncPipe()
	defer func() {
		client.Close()
		server.Close()
	}()

	// config
	tracker := HTTPTracker{}
	config.Config.DB.Type = "gomap"
	config.Config.DB.Backup.Type = "none"
	config.Config.Announce.Fuzz = 1 * time.Second
	config.Config.Numwant.Limit = 200 // for peerlistpool

	// setup db
	db, err := storage.Open()
	if err != nil {
		b.Error("failed to open storage", err)
		b.FailNow()
	}
	tracker.peerdb = db

	// setup params
	params := announceParams{
		compact:  false,
		nopeerid: false,
		noneleft: false,
		event:    "started",
		port:     "6969",
		hash:     "01234567890123456789",
		peerid:   "01234567890123456789",
		numwant:  "200",
	}
	addr := netip.MustParseAddr("123.123.123.123")

	// init entries in database
	for i := 0; i < 200; i++ {
		params.peerid = randString(20)
		tracker.announce(client, &params, addr)
	}

	b.ResetTimer()

	// run benchmark
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		params.peerid = randString(20)
		b.StartTimer()
		tracker.announce(client, &params, addr)
	}
}
