package tracker_test

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

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

func Request(infoHash, ip, event, left, peerID, key, port string) error {
	// Make the request
	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/announce", nil)
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
	spew.Dump(decoded)

	return nil
}

func TestApp(t *testing.T) {

	// Make peers
	Request("QWERTYUIOPASDFGHJKLZ", "1.1.1.1", "started", "100", "peer1", "peer1", "42069")
	Request("QWERTYUIOPASDFGHJKLZ", "2.2.2.2", "started", "100", "peer2", "peer2", "42069")

	// Update peers
	Request("QWERTYUIOPASDFGHJKLZ", "1.1.1.11", "started", "80", "peer1", "peer1", "6999")

	// Complete peers
	Request("QWERTYUIOPASDFGHJKLZ", "2.2.2.2", "started", "0", "peer1", "peer1", "6999")
	Request("QWERTYUIOPASDFGHJKLZ", "1.1.1.11", "complete", "200", "peer1", "peer1", "6999")

	return
}

func BenchmarkApp(b *testing.B) {
	rand.Seed(time.Now().Unix())
	for n := 0; n < b.N; n++ {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/announce", nil)
		q := req.URL.Query()
		q.Add("info_hash", "QWERTYUIOPASDFGHJKLZ")
		q.Add("ip", "69.69.69.69")
		q.Add("event", "started")
		q.Add("left", "100")
		q.Add("peer_id", randStr(20)) // random
		q.Add("key", randStr(20))     // random
		q.Add("port", "69")
		req.URL.RawQuery = q.Encode()

		// Send it
		client.Do(req)
	}
}
