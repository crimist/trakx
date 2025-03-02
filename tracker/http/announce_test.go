package http

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"testing"

	_ "github.com/crimist/trakx/storage/inmemory"
)

// go build -gcflags '-m' -o /dev/null ./... |& grep "moved to heap:"

func TestAnnounce(t *testing.T) {
	// rand.Seed deprecated, need GODEBUG to enforce deterministic rand
	os.Setenv("GODEBUG", "randseednop=0")
	rand.Seed(1)

	var cases = []struct {
		name           string
		ipv6           bool
		validResponses [][]byte
		// announce params
		infoHash   string
		peerId     string
		port       string
		downloaded string
		left       string
		uploaded   string
		event      string
		numWant    string
		compact    bool
		noPeerId   bool
	}{
		{
			name:       "full",
			infoHash:   "00000000000000000001",
			peerId:     "11111111111111111111",
			port:       "1234",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    false,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("d8:intervali10e8:completei0e10:incompletei1e5:peersl61:d7:peer id20:111111111111111111112:ip9:127.0.0.14:porti1234eeee"),
			},
		},
		{
			name:       "fullMulti",
			infoHash:   "00000000000000000001",
			peerId:     "22222222222222222222",
			port:       "4321",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    false,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("d8:intervali10e8:completei0e10:incompletei2e5:peersl61:d7:peer id20:111111111111111111112:ip9:127.0.0.14:porti1234ee61:d7:peer id20:222222222222222222222:ip9:127.0.0.14:porti4321eeee"),
				[]byte("d8:intervali10e8:completei0e10:incompletei2e5:peersl61:d7:peer id20:222222222222222222222:ip9:127.0.0.14:porti4321ee61:d7:peer id20:111111111111111111112:ip9:127.0.0.14:porti1234eeee"),
			},
		},
		{
			name:       "fullIPv6",
			ipv6:       true,
			infoHash:   "00000000000000000002",
			peerId:     "11111111111111111111",
			port:       "1234",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    false,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peersl58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234eeee"),
			},
		},
		{
			name:       "fullIPv6Multi",
			ipv6:       true,
			infoHash:   "00000000000000000002",
			peerId:     "22222222222222222222",
			port:       "4321",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    false,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234ee58:d7:peer id20:222222222222222222222:ip6:::56784:porti4321eeee"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl58:d7:peer id20:222222222222222222222:ip6:::56784:porti4321ee58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234eeee"),
			},
		},
		{
			name:       "fullMixedMulti",
			ipv6:       false,
			infoHash:   "00000000000000000002",
			peerId:     "22222222222222222222",
			port:       "4321",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    false,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234ee59:d7:peer id20:222222222222222222222:ip7:1.1.1.14:porti4321eeee"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl59:d7:peer id20:222222222222222222222:ip7:1.1.1.14:porti4321ee58:d7:peer id20:111111111111111111112:ip6:::12344:porti1234eeee"),
			},
		},
		{
			name:       "nopeerid",
			ipv6:       false,
			infoHash:   "11111111111111111111",
			peerId:     "11111111111111111111",
			port:       "1234",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    false,
			noPeerId:   true,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peersl27:d2:ip7:1.1.1.14:porti1234eeee"),
			},
		},
		{
			name:       "nopeeridMulti",
			ipv6:       false,
			infoHash:   "11111111111111111111",
			peerId:     "22222222222222222222",
			port:       "4321",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    false,
			noPeerId:   true,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl27:d2:ip7:1.1.1.14:porti1234ee27:d2:ip7:2.2.2.24:porti4321eeee"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peersl27:d2:ip7:2.2.2.24:porti4321ee27:d2:ip7:1.1.1.14:porti1234eeee"),
			},
		},
		{
			name:       "compact",
			ipv6:       false,
			infoHash:   "22222222222222222222",
			peerId:     "11111111111111111111",
			port:       "1234",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    true,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peers6:\x01\x01\x01\x01\x04\xd26:peers60:e"),
			},
		},
		{
			name:       "compactMulti",
			ipv6:       false,
			infoHash:   "22222222222222222222",
			peerId:     "22222222222222222222",
			port:       "4321",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    true,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers12:\x01\x01\x01\x01\x04\xd2\x02\x02\x02\x02\x10\xe16:peers60:e"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers12:\x02\x02\x02\x02\x10\xe1\x01\x01\x01\x01\x04\xd26:peers60:e"),
			},
		},
		{
			name:       "compactIPv6",
			ipv6:       true,
			infoHash:   "33333333333333333333",
			peerId:     "11111111111111111111",
			port:       "1234",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    true,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei1e5:peers0:6:peers618:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2e"),
			},
		},
		{
			name:       "compactIPv6Multi",
			ipv6:       true,
			infoHash:   "33333333333333333333",
			peerId:     "22222222222222222222",
			port:       "1234",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    true,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers0:6:peers636:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x56\x78\x04\xd2e"),
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers0:6:peers636:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x56\x78\x04\xd2\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2e"),
			},
		},
		{
			name:       "compactMixedMulti",
			ipv6:       false,
			infoHash:   "33333333333333333333",
			peerId:     "22222222222222222222",
			port:       "1234",
			downloaded: "0",
			left:       "1000",
			uploaded:   "0",
			event:      "started",
			numWant:    "10",
			compact:    true,
			noPeerId:   false,
			validResponses: [][]byte{
				[]byte("HTTP/1.1 200\r\n\r\nd8:intervali10e8:completei0e10:incompletei2e5:peers6:\x01\x01\x01\x01\x04\xd26:peers618:\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x12\x34\x04\xd2e"),
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			params := url.Values{}
			params.Add("info_hash", c.infoHash)
			params.Add("peer_id", c.peerId)
			params.Add("port", c.port)
			params.Add("downloaded", c.downloaded)
			params.Add("left", c.left)
			params.Add("uploaded", c.uploaded)
			params.Add("event", c.event)
			params.Add("numwant", c.numWant)

			if c.compact {
				params.Add("compact", "1")
			}
			if c.noPeerId {
				params.Add("no_peer_id", "1")
			}

			serverAddress := testNetAddress4
			if c.ipv6 {
				serverAddress = testNetAddress6
			}

			announceURL := fmt.Sprintf("http://%s:%d/announce?%s", serverAddress, testNetworkPort, params.Encode())

			resp, err := http.Get(announceURL)
			if err != nil {
				t.Fatal("failed to http get", err)
			}

			if resp.StatusCode != 200 {
				t.Fatalf("Expected code 200, got %d", resp.StatusCode)
			}

			body := make([]byte, 65535)
			n, err := resp.Body.Read(body)
			if err != nil {
				t.Fatal("failed to read from connection", err)
			}
			body = body[:n]

			for _, expectedResp := range c.validResponses {
				if bytes.Equal(expectedResp, body) {
					return
				}
			}

			var expectedDump string
			for _, expectedResp := range c.validResponses {
				expectedDump += "\n" + hex.Dump(expectedResp)
			}

			t.Errorf("bad announce for %v\nresp:\n%v\nexpected:%v", c.name, hex.Dump(body), expectedDump)
		})
	}
}
