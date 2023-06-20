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

func BenchmarkAnnounce(b *testing.B) {
	// setup pipe
	conn := connutil.Discard()

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
	for i := 0; i < 300; i++ {
		params.peerid = randString(20)
		tracker.announce(conn, &params, addr)
	}

	b.ResetTimer()

	// run benchmark
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		params.peerid = randString(20)
		b.StartTimer()
		tracker.announce(conn, &params, addr)
	}
}

func BenchmarkAnnounceCompact(b *testing.B) {
	// setup pipe
	conn := connutil.Discard()

	// config
	tracker := HTTPTracker{}
	config.Config.DB.Type = "gomap"
	config.Config.DB.Backup.Type = "none"
	config.Config.Announce.Fuzz = 1 * time.Second
	config.Config.Numwant.Limit = 200

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
		tracker.announce(conn, &params, addr)
	}

	b.ResetTimer()

	// run benchmark
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		params.peerid = randString(20)
		b.StartTimer()
		tracker.announce(conn, &params, addr)
	}
}

func BenchmarkAnnounceCompactParallel(b *testing.B) {
	// setup pipe
	conn := connutil.Discard()

	// config
	tracker := HTTPTracker{}
	config.Config.DB.Type = "gomap"
	config.Config.DB.Backup.Type = "none"
	config.Config.Announce.Fuzz = 1 * time.Second
	config.Config.Numwant.Limit = 200

	// setup db
	db, err := storage.Open()
	if err != nil {
		b.Error("failed to open storage", err)
		b.FailNow()
	}
	tracker.peerdb = db

	// setup setupParams
	setupParams := announceParams{
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
		setupParams.peerid = randString(20)
		tracker.announce(conn, &setupParams, addr)
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			params := announceParams{
				compact:  true,
				nopeerid: true,
				noneleft: false,
				event:    "started",
				port:     "6969",
				hash:     "01234567890123456789",
				peerid:   randString(20),
				numwant:  "200",
			}
			tracker.announce(conn, &params, addr)
		}
	})
}
