package http

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"net"
	"net/netip"
	"runtime/debug"
	"testing"
	"time"

	"github.com/cbeuw/connutil"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"

	_ "github.com/crimist/trakx/tracker/storage/map"
)

// go build -gcflags '-m' -o /dev/null ./... |& grep "moved to heap:"

func TestHTTPAnnounce(t *testing.T) {
	rand.Seed(1) // golang default

	// setup config
	config.Conf.DB.Type = "gomap"
	config.Conf.DB.Backup.Type = "none"
	config.Conf.Announce.Base = 10 * time.Second
	config.Conf.Announce.Fuzz = 0

	// setup db
	db, err := storage.Open()
	if err != nil {
		t.Error("failed to open storage", err)
		t.FailNow()
	}

	// setup tracker
	tracker := HTTPTracker{}
	tracker.peerdb = db

	// setup pipe
	client, server := connutil.AsyncPipe()
	defer func() {
		client.Close()
		server.Close()
	}()

	var cases = []struct {
		name              string
		params            announceParams
		ip                netip.Addr
		expectedResponses [][]byte
	}{
		{
			"full",
			announceParams{
				compact:  false,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "1234",
				hash:     "00000000000000000001",
				peerid:   "11111111111111111111",
				numwant:  "10",
			},
			netip.MustParseAddr("1.1.1.1"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peersl59:d7:peer id20:111111111111111111112:ip7:1.1.1.14:porti1234eeee"),
			},
		},
		{
			"fullMulti",
			announceParams{
				compact:  false,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "4321",
				hash:     "00000000000000000001",
				peerid:   "22222222222222222222",
				numwant:  "10",
			},
			netip.MustParseAddr("2.2.2.2"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl59:d7:peer id20:111111111111111111112:ip7:1.1.1.14:porti1234ee59:d7:peer id20:222222222222222222222:ip7:2.2.2.24:porti4321eeee"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl59:d7:peer id20:222222222222222222222:ip7:2.2.2.24:porti4321ee59:d7:peer id20:111111111111111111112:ip7:1.1.1.14:porti1234eeee"),
			},
		},
		{
			"fullIPv6",
			announceParams{
				compact:  false,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "1234",
				hash:     "00000000000000000002",
				peerid:   "11111111111111111111",
				numwant:  "10",
			},
			netip.MustParseAddr("::1234"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peersl58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234eeee"),
			},
		},
		{
			"fullIPv6Multi",
			announceParams{
				compact:  false,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "4321",
				hash:     "00000000000000000002",
				peerid:   "22222222222222222222",
				numwant:  "10",
			},
			netip.MustParseAddr("::5678"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234ee58:d7:peer id20:222222222222222222222:ip6:::56784:porti4321eeee"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl58:d7:peer id20:222222222222222222222:ip6:::56784:porti4321ee58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234eeee"),
			},
		},
		{
			"fullMixedMulti",
			announceParams{
				compact:  false,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "4321",
				hash:     "00000000000000000002",
				peerid:   "22222222222222222222",
				numwant:  "10",
			},
			netip.MustParseAddr("1.1.1.1"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234ee59:d7:peer id20:222222222222222222222:ip7:1.1.1.14:porti4321eeee"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl59:d7:peer id20:222222222222222222222:ip7:1.1.1.14:porti4321ee58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234eeee"),
			},
		},
		{
			"nopeerid",
			announceParams{
				compact:  false,
				nopeerid: true,
				noneleft: false,
				event:    "started",
				port:     "1234",
				hash:     "11111111111111111111",
				peerid:   "11111111111111111111",
				numwant:  "10",
			},
			netip.MustParseAddr("1.1.1.1"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peersl27:d2:ip7:1.1.1.14:porti1234eeee"),
			},
		},
		{
			"nopeeridMulti",
			announceParams{
				compact:  false,
				nopeerid: true,
				noneleft: false,
				event:    "started",
				port:     "4321",
				hash:     "11111111111111111111",
				peerid:   "22222222222222222222",
				numwant:  "10",
			},
			netip.MustParseAddr("2.2.2.2"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl27:d2:ip7:1.1.1.14:porti1234ee27:d2:ip7:2.2.2.24:porti4321eeee"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl27:d2:ip7:2.2.2.24:porti4321ee27:d2:ip7:1.1.1.14:porti1234eeee"),
			},
		},
		{
			"compact",
			announceParams{
				compact:  true,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "1234",
				hash:     "22222222222222222222",
				peerid:   "11111111111111111111",
				numwant:  "10",
			},
			netip.MustParseAddr("1.1.1.1"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peers6:\x01\x01\x01\x01\x04\xd26:peers60:e"),
			},
		},
		{
			"compactMulti",
			announceParams{
				compact:  true,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "4321",
				hash:     "22222222222222222222",
				peerid:   "22222222222222222222",
				numwant:  "10",
			},
			netip.MustParseAddr("2.2.2.2"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers12:\x01\x01\x01\x01\x04\xd2\x02\x02\x02\x02\x10\xe16:peers60:e"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers12:\x02\x02\x02\x02\x10\xe1\x01\x01\x01\x01\x04\xd26:peers60:e"),
			},
		},
		{
			"compactIPv6",
			announceParams{
				compact:  true,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "1234",
				hash:     "33333333333333333333",
				peerid:   "11111111111111111111",
				numwant:  "10",
			},
			netip.MustParseAddr("::1234"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peers0:6:peers618:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2e"),
			},
		},
		{
			"compactIPv6Multi",
			announceParams{
				compact:  true,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "1234",
				hash:     "33333333333333333333",
				peerid:   "22222222222222222222",
				numwant:  "10",
			},
			netip.MustParseAddr("::5678"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers0:6:peers636:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x56\x78\x04\xd2e"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers0:6:peers636:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x56\x78\x04\xd2\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2e"),
			},
		},
		{
			"compactMixedMulti",
			announceParams{
				compact:  true,
				nopeerid: false,
				noneleft: false,
				event:    "started",
				port:     "1234",
				hash:     "33333333333333333333",
				peerid:   "22222222222222222222",
				numwant:  "10",
			},
			netip.MustParseAddr("1.1.1.1"),
			[][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers6:\x01\x01\x01\x01\x04\xd26:peers618:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2e"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp := make([]byte, 0xFFFF)

			tracker.announce(client, &c.params, c.ip)
			respSize, err := server.Read(resp)
			if err != nil {
				t.Error("Error reading asyncpipe")
			}
			resp = resp[:respSize]

			for _, expectedResp := range c.expectedResponses {
				if bytes.Equal(expectedResp, resp) {
					return
				}
			}

			var expectedDump string
			for _, expectedResp := range c.expectedResponses {
				expectedDump += "\n" + hex.Dump(expectedResp)
			}

			t.Errorf("bad announce for %v\nresp:\n%v\nexpected:%v", c.name, hex.Dump(resp), expectedDump)
		})
	}
}

// TODO: clean up these announce benchmarks

const (
	addr = ":12345"
)

func BenchmarkHTTPAnnounceCompact200(b *testing.B) {
	b.Log("This benchmark does not reflect the real memory usage of Announce. However it can be used for comparative purposes")

	// open listen
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	// run accepter & reader
	go func() {
		date := make([]byte, 65535)
		c, _ := ln.Accept()
		for {
			if _, err := c.Read(date); err != nil {
				break
			}
		}
	}()

	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		panic(err)
	}

	// setup tracker
	tracker := HTTPTracker{}
	config.Conf.DB.Type = "gomap"
	config.Conf.DB.Backup.Type = "none"
	config.Conf.Announce.Fuzz = 1 * time.Second
	config.Conf.Numwant.Limit = 300 // for peerlistpool

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
	ip := netip.MustParseAddr("123.123.123.123")

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
		b.StopTimer()
		params.peerid = random20()
		b.StartTimer()
		tracker.announce(conn, &params, ip)
	}

	// cleanup
	b.StopTimer()
	debug.SetGCPercent(gcp)
}

func BenchmarkHTTPAnnounce200(b *testing.B) {
	b.Log("This benchmark does not reflect the real memory usage of Announce. However it can be used for comparative purposes")

	// open listen
	ln, err := net.Listen("tcp4", addr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	// run accepter & reader
	go func() {
		date := make([]byte, 65535)
		c, _ := ln.Accept()
		for {
			if _, err := c.Read(date); err != nil {
				break
			}
		}
	}()

	conn, err := net.Dial("tcp4", addr)
	if err != nil {
		panic(err)
	}

	// setup tracker
	tracker := HTTPTracker{}
	config.Conf.DB.Type = "gomap"
	config.Conf.DB.Backup.Type = "none"
	config.Conf.Announce.Fuzz = 1 * time.Second
	config.Conf.Numwant.Limit = 300 // for peerlistpool

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
	ip := netip.MustParseAddr("123.123.123.123")

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
		b.StopTimer()
		params.peerid = random20()
		b.StartTimer()
		tracker.announce(conn, &params, ip)
	}

	// cleanup
	b.StopTimer()
	debug.SetGCPercent(gcp)
}
