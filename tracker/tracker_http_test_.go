package tracker

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-torrent/bencode"
)

const (
	announceHTTPaddress = "http://127.0.0.1:1337/announce"
)

var httpClient = &http.Client{
	Timeout: 500 * time.Millisecond,
}

func TestHTTPAnnounce(t *testing.T) {
	req, err := http.NewRequest("GET", announceHTTPaddress, nil)
	if err != nil {
		t.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("info_hash", "TestAnnounceHTTP1234")
	q.Add("event", "started")
	q.Add("left", "1000")
	q.Add("downloaded", "0")
	q.Add("peer_id", "QB123456789012345678")
	q.Add("key", "useless")
	q.Add("port", "1234")
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Skip("HTTP tracker not running")
		}
		t.Fatal("HTTP request error:", err)
	}

	// Parse it
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Body read error:", err)
	}
	resp.Body.Close()

	if len(body) == 0 {
		t.Fatal("body empty")
	}

	var decoded map[string]interface{}
	if err = bencode.Unmarshal(body, &decoded); err != nil {
		t.Fatal("Unmarshalling error:", err)
	}

	if _, ok := decoded["failure reason"]; ok {
		t.Error("Tracker error:", decoded["failure reason"])
	}

	if decoded["complete"] != 0 {
		t.Error("Num complete should be 0 got,", decoded["complete"])
	}
	if decoded["incomplete"] != 1 {
		t.Error("Num incomplete should be 1 got,", decoded["incomplete"])
	}
	var peer map[string]interface{}
	if err = bencode.Unmarshal([]byte(decoded["peers"].(bencode.List)[0].(string)), &peer); err != nil {
		t.Fatal(err)
	}
	if peer["peer id"] != "QB123456789012345678" {
		t.Error("PeerID should be QB123456789012345678 got,", peer["peer id"])
	}
	if peer["ip"] != "127.0.0.1" {
		t.Error("ip should be 127.0.0.1 got,", peer["ip"])
	}
	if peer["port"] != 1234 {
		t.Error("port should be 1234 got,", peer["port"])
	}
}

func TestHTTPAnnounceCompact(t *testing.T) {
	req, err := http.NewRequest("GET", announceHTTPaddress, nil)
	if err != nil {
		t.Error(err)
	}

	q := req.URL.Query()
	q.Add("info_hash", "TestAnnounceCompactH")
	q.Add("event", "started")
	q.Add("left", "1000")
	q.Add("downloaded", "0")
	q.Add("peer_id", "QB123456789012345678")
	q.Add("key", "useless")
	q.Add("port", "1234")
	q.Add("compact", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Skip("HTTP tracker not running")
		}
		t.Fatal("HTTP request error:", err)
	}

	// Parse it
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Body read error:", err)
	}
	resp.Body.Close()
	var decoded map[string]interface{}
	if err = bencode.Unmarshal(body, &decoded); err != nil {
		t.Fatal("Unmarshalling error:", err)
	}

	if _, ok := decoded["failure reason"]; ok {
		t.Error("Tracker error:", decoded["failure reason"])
	}

	peerBytes := []byte(decoded["peers"].(string))
	if len(peerBytes) != 6 {
		t.Fatal("len(peers) should be 6, got", len(peerBytes))
	}
	if !bytes.Equal(peerBytes[0:4], []byte{127, 0, 0, 1}) {
		t.Error("ip should be [127, 0, 0, 1], got", peerBytes[4:6])
	}
	if !bytes.Equal(peerBytes[4:6], []byte{0x04, 0xD2}) {
		t.Error("port should be [4, 210], got", peerBytes[4:6])
	}
}
