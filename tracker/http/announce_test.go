package http

import (
	"bytes"
	"encoding/hex"
	"math/rand"
	"net/netip"
	"testing"
	"time"

	"github.com/cbeuw/connutil"
	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/storage"

	_ "github.com/crimist/trakx/tracker/storage/map"
)

// go build -gcflags '-m' -o /dev/null ./... |& grep "moved to heap:"

func TestAnnounce(t *testing.T) {
	rand.Seed(1) // golang default

	// setup config
	config.Config.DB.Type = "gomap"
	config.Config.DB.Backup.Type = "none"
	config.Config.Announce.Base = 10 * time.Second
	config.Config.Announce.Fuzz = 0
	config.Config.Numwant.Limit = 10

	// setup pools
	pools.Initialize(10)

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
