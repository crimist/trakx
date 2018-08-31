package tracker_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/go-torrent/bencode"
)

var client = &http.Client{}

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
	// Makes fake peers n shit

	Request("QWERTYUIOPASDFGHJKLZ", "1.1.1.1", "started", "100", "peer1", "peer1", "42069")
	Request("QWERTYUIOPASDFGHJKLZ", "2.2.2.2", "started", "100", "peer2", "peer2", "42069")

	return
}
