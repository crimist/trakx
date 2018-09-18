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
	fmt.Println("--------------------------------------")
	spew.Dump(peerID, decoded, resp.StatusCode)

	switch v := decoded["peers"].(type) {
	case string:
		spew.Dump([]byte(v))
	default:
	}

	return nil
}

func TestApp(t *testing.T) {

	// Make peers
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.1", "started", "100", "AAAAAAAAAAAAAAAPEER1", "peer1", "42069", false)
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.2", "started", "100", "AAAAAAAAAAAAAAAPEER2", "peer2", "42069", false)

	// Update peers
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.11", "started", "80", "AAAAAAAAAAAAAAAPEER1", "peer1", "6999", false)

	// Complete peers
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.2", "started", "0", "AAAAAAAAAAAAAAAPEER1", "peer1", "6999", false)
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.11", "completed", "200", "AAAAAAAAAAAAAAAPEER1", "peer1", "6999", false)

	// Compact responses
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.3", "started", "0", "AAAAAAAAAAAAAAAPEER3", "peer3", "4213", true)

	// Ipv6
	Request("QWERTYUIOPASDFGHJKLZ", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "started", "0", "AAAAAAAAAAAAAAAPEER4", "peer4", "8765", true)

	// Test banned hashes
	Request("8C4947E96C7C9F770AA3", "192.168.1.4", "started", "0", "AAAAAAAAAAAAAAAPEER5", "peer5", "1111", false)

	// Remove them all
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.1", "stopped", "100", "AAAAAAAAAAAAAAAPEER1", "peer1", "42069", false)
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.2", "stopped", "100", "AAAAAAAAAAAAAAAPEER2", "peer2", "42069", false)
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.11", "stopped", "80", "AAAAAAAAAAAAAAAPEER1", "peer1", "6999", false)
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.2", "stopped", "0", "AAAAAAAAAAAAAAAPEER1", "peer1", "6999", false)
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.11", "stopped", "200", "AAAAAAAAAAAAAAAPEER1", "peer1", "6999", false)
	Request("QWERTYUIOPASDFGHJKLZ", "192.168.1.3", "stopped", "0", "AAAAAAAAAAAAAAAPEER3", "peer3", "4213", true)
	Request("QWERTYUIOPASDFGHJKLZ", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "stopped", "0", "AAAAAAAAAAAAAAAPEER4", "peer4", "8765", true)

	return
}
