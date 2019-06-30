package tracker_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/go-torrent/bencode"
)

var client = &http.Client{}

func TestAnnounce(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:1337/announce", nil)
	if err != nil {
		t.Error(err)
	}

	q := req.URL.Query()
	q.Add("info_hash", "TestAnnounce12345678")
	q.Add("ip", "123.123.123.123")
	q.Add("event", "started")
	q.Add("left", "1000")
	q.Add("downloaded", "0")
	q.Add("peer_id", "QB123456789012345678")
	q.Add("key", "useless")
	q.Add("port", "1234")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	// Parse it
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	resp.Body.Close()
	var decoded map[string]interface{}
	if err = bencode.Unmarshal(body, &decoded); err != nil {
		t.Error(err)
	}

	if _, ok := decoded["failure reason"]; ok {
		t.Error("Tracker error:", decoded["failure reason"])
	}

	if decoded["complete"] != 0 {
		t.Error("Num complete should be 0 got", decoded["complete"])
	}
	if decoded["incomplete"] != 1 {
		t.Error("Num incomplete should be 1 got", decoded["incomplete"])
	}
	var peer map[string]interface{}
	if err = bencode.Unmarshal([]byte(decoded["peers"].(bencode.List)[0].(string)), &peer); err != nil {
		t.Error(err)
	}
	if peer["peer id"] != "QB123456789012345678" {
		t.Error("PeerID should be QB123456789012345678 got", peer["peer id"])
	}
	if peer["ip"] != "123.123.123.123" {
		t.Error("ip should be 123.123.123.123 got", peer["ip"])
	}
	if peer["port"] != 1234 {
		t.Error("port should be 1234 got", peer["port"])
	}
}

func TestAnnounceCompact(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:1337/announce", nil)
	if err != nil {
		t.Error(err)
	}

	q := req.URL.Query()
	q.Add("info_hash", "TestAnnounceCompact1")
	q.Add("ip", "123.123.123.123")
	q.Add("event", "started")
	q.Add("left", "1000")
	q.Add("downloaded", "0")
	q.Add("peer_id", "QB123456789012345678")
	q.Add("key", "useless")
	q.Add("port", "1234")
	q.Add("compact", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}

	// Parse it
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	resp.Body.Close()
	var decoded map[string]interface{}
	if err = bencode.Unmarshal(body, &decoded); err != nil {
		t.Error(err)
	}

	if _, ok := decoded["failure reason"]; ok {
		t.Error("Tracker error:", decoded["failure reason"])
	}

	peerBytes := []byte(decoded["peers"].(string))
	if len(peerBytes) != 6 {
		t.Error("len(peers) should be 6 got", len(peerBytes))
	}
	if bytes.Compare(peerBytes[0:4], []byte{0x7B, 0x7B, 0x7B, 0x7B}) != 0 {
		t.Error("ip should be [7B, 7B, 7B, 7B] got", peerBytes[4:6])
	}
	if bytes.Compare(peerBytes[4:6], []byte{0x04, 0xD2}) != 0 {
		t.Error("port should be [4, 210] got", peerBytes[4:6])
	}
}
