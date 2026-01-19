package http

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/go-torrent/bencode"
	"github.com/pkg/errors"
)

type torrentStats struct {
	Complete   int
	Incomplete int
}

func sendScrape(address string, params url.Values) (map[string]torrentStats, error) {
	announceURL := fmt.Sprintf("http://%s:%d/scrape?%s", address, testNetworkPort, params.Encode())

	resp, err := http.Get(announceURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get url")
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("non 200 status code: " + resp.Status)
	}

	body := make([]byte, 65535)
	n, err := resp.Body.Read(body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}
	body = body[:n]

	var root map[string]interface{}
	if err = bencode.Unmarshal(body, &root); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal bencoded body")
	}

	if _, ok := root["failure reason"]; ok {
		return nil, errors.New("tracker error: " + root["failure reason"].(string))
	}

	torrents := make(map[string]torrentStats)
	for hash, stats := range root["files"].(bencode.Dictionary) {
		torrents[hash] = torrentStats{
			Complete:   stats.(bencode.Dictionary)["complete"].(int),
			Incomplete: stats.(bencode.Dictionary)["incomplete"].(int),
		}
	}

	return torrents, nil
}

func TestScrape(t *testing.T) {
	params := url.Values{}
	params.Add("info_hash", "00000000000000000006")
	params.Add("peer_id", "00000000000000000001")
	params.Add("port", "1")
	params.Add("downloaded", "0")
	params.Add("left", "1000")
	params.Add("uploaded", "0")
	params.Add("event", "started")
	_, _, err := sendAnnounce(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	params = url.Values{}
	params.Add("info_hash", "00000000000000000006")
	torrents, err := sendScrape(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	if len(torrents) != 1 {
		t.Errorf("Expected torrents len = 1; got %v", len(torrents))
	}
	if torrents["00000000000000000006"].Complete != 0 {
		t.Errorf("Expected complete = 0; got %v", torrents["00000000000000000006"].Complete)
	}
	if torrents["00000000000000000006"].Incomplete != 1 {
		t.Errorf("Expected incomplete = 1; got %v", torrents["00000000000000000006"].Incomplete)
	}

	params.Add("info_hash", "00000000000000000007")
	torrents, err = sendScrape(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	if len(torrents) != 2 {
		t.Errorf("Expected torrents len = 2; got %v", len(torrents))
	}
	if torrents["00000000000000000006"].Complete != 0 {
		t.Errorf("Expected complete = 0; got %v", torrents["00000000000000000006"].Complete)
	}
	if torrents["00000000000000000006"].Incomplete != 1 {
		t.Errorf("Expected incomplete = 1; got %v", torrents["00000000000000000006"].Incomplete)
	}
	if torrents["00000000000000000007"].Complete != 0 {
		t.Errorf("Expected complete = 0; got %v", torrents["00000000000000000007"].Complete)
	}
	if torrents["00000000000000000007"].Incomplete != 0 {
		t.Errorf("Expected incomplete = 0; got %v", torrents["00000000000000000007"].Incomplete)
	}
}
