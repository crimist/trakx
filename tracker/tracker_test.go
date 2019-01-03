package tracker_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/go-torrent/bencode"
)

var client = &http.Client{}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()-_=+[{]}\\|\":;<,>.?'/~`"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func randStr(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func Request(infoHash, ip, event, left, peerID, key, port string, compact bool) error {
	// Make the request
	req, err := http.NewRequest("GET", "http://127.0.0.1:1337/announce", nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Add("info_hash", infoHash)
	q.Add("ip", ip)
	q.Add("event", event)
	q.Add("left", left)
	q.Add("peer_id", peerID)
	q.Add("key", key)
	q.Add("port", port)
	if compact {
		q.Add("compact", "1")
	}
	req.URL.RawQuery = q.Encode()

	// Send it
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Parse it
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var decoded map[string]interface{}
	bencode.Unmarshal(body, &decoded)
	fmt.Println("\n--------------------------------------")
	spew.Dump(peerID, decoded, resp.StatusCode)

	switch v := decoded["peers"].(type) {
	case string:
		spew.Dump([]byte(v))
	default:
	}

	return nil
}

func TestApp(t *testing.T) {
	// t.Skip()

	// Make peers
	Request("ABCDEFGHIJKLMNOPQRST", "1.1.1.1", "started", "100", "PEER1_______________", "peer1", "8000", false)
	Request("ABCDEFGHIJKLMNOPQRST", "2.2.2.2", "started", "100", "PEER2_______________", "peer2", "8000", false)

	// Update peers
	Request("ABCDEFGHIJKLMNOPQRST", "11.11.11.11", "started", "50", "PEER1_______________", "peer1", "8888", false)

	// Complete peers
	Request("ABCDEFGHIJKLMNOPQRST", "11.11.11.11", "started", "0", "PEER1_______________", "peer1", "8888", false)
	Request("ABCDEFGHIJKLMNOPQRST", "192.168.1.11", "completed", "0", "PEER2_______________", "peer2", "8080", false)

	// Compact responses
	Request("ABCDEFGHIJKLMNOPQRST", "192.168.1.3", "started", "0", "PEER3_______________", "peer3", "8080", true)

	// Ipv6
	Request("ABCDEFGHIJKLMNOPQRST", "::1", "started", "100", "PEER4_______________", "peer4", "8080", false)

	// Test banned hashes
	Request("8C4947E96C7C9F770AA3", "192.168.1.4", "started", "100", "PEER5_______________", "peer5", "1111", false)

	return
}
