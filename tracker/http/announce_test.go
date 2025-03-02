package http

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"testing"

	_ "github.com/crimist/trakx/storage/inmemory"
	"github.com/go-torrent/bencode"
	"github.com/pkg/errors"
)

// go build -gcflags '-m' -o /dev/null ./... |& grep "moved to heap:"

// sendAnnounce returns the bencoded response root and a list of peers, handles compact and non-compact responses
func sendAnnounce(address string, params url.Values) (map[string]interface{}, []map[string]interface{}, error) {
	announceURL := fmt.Sprintf("http://%s:%d/announce?%s", address, testNetworkPort, params.Encode())

	resp, err := http.Get(announceURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to get url")
	}

	if resp.StatusCode != 200 {
		return nil, nil, errors.New("non 200 status code")
	}

	body := make([]byte, 65535)
	n, err := resp.Body.Read(body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to read response body")
	}
	body = body[:n]

	var root map[string]interface{}
	if err = bencode.Unmarshal(body, &root); err != nil {
		return nil, nil, errors.Wrap(err, "failed to unmarshal bencoded body")
	}

	if _, ok := root["failure reason"]; ok {
		return nil, nil, errors.New("tracker error: " + root["failure reason"].(string))
	}

	dictPeers := root["peers"]
	var peers []map[string]interface{}

	switch dictPeers := dictPeers.(type) {
	case bencode.List:
		for _, peer := range dictPeers {
			var p map[string]interface{}
			if err := bencode.Unmarshal([]byte(peer.(string)), &p); err != nil {
				return nil, nil, errors.Wrap(err, "failed to unmarshal peer")
			}
			peers = append(peers, p)
		}
	case string:
		for i := 0; i < len(dictPeers); i += 6 {
			peers = append(peers, map[string]interface{}{
				"ip":   fmt.Sprintf("%d.%d.%d.%d", dictPeers[i], dictPeers[i+1], dictPeers[i+2], dictPeers[i+3]),
				"port": int(dictPeers[i+4])<<8 | int(dictPeers[i+5]),
			})
		}
	default:
		return nil, nil, errors.New("peers is not a list")
	}

	if _, ok := root["peers6"]; ok {
		dictPeers6 := root["peers6"].(string)
		for i := 0; i < len(dictPeers6); i += 18 {
			peers = append(peers, map[string]interface{}{
				"ip":   net.IP(dictPeers6[i : i+16]).String(),
				"port": int(dictPeers6[i+16])<<8 | int(dictPeers6[i+17]),
			})
		}
	}

	return root, peers, nil
}

func TestAnnounce4(t *testing.T) {
	params := url.Values{}
	params.Add("info_hash", "00000000000000000001")
	params.Add("peer_id", "00000000000000000001")
	params.Add("port", "1")
	params.Add("downloaded", "0")
	params.Add("left", "1000")
	params.Add("uploaded", "0")
	params.Add("event", "started")
	root, peers, err := sendAnnounce(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	if root["interval"] != 10 {
		t.Errorf("Expected interval = %v; got %v", 10, root["interval"])
	}
	if root["complete"] != 0 {
		t.Errorf("Expected complete = %v; got %v", 0, root["complete"])
	}
	if root["incomplete"] != 1 {
		t.Errorf("Expected incomplete = %v; got %v", 1, root["incomplete"])
	}
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer; got %d", len(peers))
	}
	if peers[0]["peer id"] != "00000000000000000001" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000001", peers[0]["peer id"])
	}
	if peers[0]["ip"] != "127.0.0.1" {
		t.Errorf("Expected ip = %v; got %v", "127.0.0.1", peers[0]["ip"])
	}
	if peers[0]["port"] != 1 {
		t.Errorf("Expected port = %v; got %v", 1, peers[0]["port"])
	}

	params.Set("peer_id", "00000000000000000002")
	_, peers, err = sendAnnounce(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	for _, peer := range peers {
		if peer["ip"] != "127.0.0.1" {
			t.Errorf("Expected ip = %v; got %v", "127.0.0.1", peer["ip"])
		}
		if peer["port"] != 1 {
			t.Errorf("Expected port = %v; got %v", 1, peer["port"])
		}
		ids = append(ids, peer["peer id"].(string))
	}
	sort.Strings(ids)
	if ids[0] != "00000000000000000001" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000001", ids[0])
	}
	if ids[1] != "00000000000000000002" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000002", ids[1])
	}
}

func TestAnnounce6(t *testing.T) {
	params := url.Values{}
	params.Add("info_hash", "00000000000000000002")
	params.Add("peer_id", "00000000000000000001")
	params.Add("port", "1")
	params.Add("downloaded", "0")
	params.Add("left", "1000")
	params.Add("uploaded", "0")
	params.Add("event", "started")
	root, peers, err := sendAnnounce(testNetAddress6, params)
	if err != nil {
		t.Fatal(err)
	}

	if root["interval"] != 10 {
		t.Errorf("Expected interval = %v; got %v", 10, root["interval"])
	}
	if root["complete"] != 0 {
		t.Errorf("Expected complete = %v; got %v", 0, root["complete"])
	}
	if root["incomplete"] != 1 {
		t.Errorf("Expected incomplete = %v; got %v", 1, root["incomplete"])
	}
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer; got %d", len(peers))
	}
	if peers[0]["peer id"] != "00000000000000000001" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000001", peers[0]["peer id"])
	}
	if peers[0]["ip"] != "::1" {
		t.Errorf("Expected ip = %v; got %v", "::1", peers[0]["ip"])
	}
	if peers[0]["port"] != 1 {
		t.Errorf("Expected port = %v; got %v", 1, peers[0]["port"])
	}

	params.Set("peer_id", "00000000000000000002")
	_, peers, err = sendAnnounce(testNetAddress6, params)
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	for _, peer := range peers {
		if peer["ip"] != "::1" {
			t.Errorf("Expected ip = %v; got %v", "::1", peer["ip"])
		}
		if peer["port"] != 1 {
			t.Errorf("Expected port = %v; got %v", 1, peer["port"])
		}
		ids = append(ids, peer["peer id"].(string))
	}
	sort.Strings(ids)
	if ids[0] != "00000000000000000001" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000001", ids[0])
	}
	if ids[1] != "00000000000000000002" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000002", ids[1])
	}
}

func TestAnnounceMixed(t *testing.T) {
	params := url.Values{}
	params.Add("info_hash", "00000000000000000003")
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

	params.Set("peer_id", "00000000000000000002")
	_, peers, err := sendAnnounce(testNetAddress6, params)
	if err != nil {
		t.Fatal(err)
	}

	var ids []string
	var ips []string
	for _, peer := range peers {
		if peer["port"] != 1 {
			t.Errorf("Expected port = %v; got %v", 1, peer["port"])
		}
		ips = append(ips, peer["ip"].(string))
		ids = append(ids, peer["peer id"].(string))
	}

	sort.Strings(ips)
	if ips[0] != "127.0.0.1" {
		t.Errorf("Expected ip = %v; got %v", "127.0.0.1", ips[0])
	}
	if ips[1] != "::1" {
		t.Errorf("Expected ip = %v; got %v", "::1", ips[1])
	}

	sort.Strings(ids)
	if ids[0] != "00000000000000000001" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000001", ids[0])
	}
	if ids[1] != "00000000000000000002" {
		t.Errorf("Expected peer id = %v; got %v", "00000000000000000002", ids[1])
	}
}

func TestAnnounceNopeerid(t *testing.T) {
	params := url.Values{}
	params.Add("info_hash", "00000000000000000004")
	params.Add("peer_id", "00000000000000000001")
	params.Add("port", "1")
	params.Add("downloaded", "0")
	params.Add("left", "1000")
	params.Add("uploaded", "0")
	params.Add("event", "started")
	params.Add("no_peer_id", "1")
	root, peers, err := sendAnnounce(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	if root["interval"] != 10 {
		t.Errorf("Expected interval = %v; got %v", 10, root["interval"])
	}
	if root["complete"] != 0 {
		t.Errorf("Expected complete = %v; got %v", 0, root["complete"])
	}
	if root["incomplete"] != 1 {
		t.Errorf("Expected incomplete = %v; got %v", 1, root["incomplete"])
	}
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer; got %d", len(peers))
	}
	if _, ok := peers[0]["peer id"]; ok {
		t.Errorf("Expected no peer id; got %v", peers[0]["peer id"])
	}
	if peers[0]["ip"] != "127.0.0.1" {
		t.Errorf("Expected ip = %v; got %v", "127.0.0.1", peers[0]["ip"])
	}
	if peers[0]["port"] != 1 {
		t.Errorf("Expected port = %v; got %v", 1, peers[0]["port"])
	}

	params.Set("peer_id", "00000000000000000002")
	_, peers, err = sendAnnounce(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	for _, peer := range peers {
		if peer["ip"] != "127.0.0.1" {
			t.Errorf("Expected ip = %v; got %v", "127.0.0.1", peer["ip"])
		}
		if peer["port"] != 1 {
			t.Errorf("Expected port = %v; got %v", 1, peer["port"])
		}
	}
}

func TestAnnounceCompact(t *testing.T) {
	params := url.Values{}
	params.Add("info_hash", "00000000000000000005")
	params.Add("peer_id", "00000000000000000001")
	params.Add("port", "1")
	params.Add("downloaded", "0")
	params.Add("left", "1000")
	params.Add("uploaded", "0")
	params.Add("event", "started")
	params.Add("compact", "1")
	root, peers, err := sendAnnounce(testNetAddress4, params)
	if err != nil {
		t.Fatal(err)
	}

	if root["interval"] != 10 {
		t.Errorf("Expected interval = %v; got %v", 10, root["interval"])
	}
	if root["complete"] != 0 {
		t.Errorf("Expected complete = %v; got %v", 0, root["complete"])
	}
	if root["incomplete"] != 1 {
		t.Errorf("Expected incomplete = %v; got %v", 1, root["incomplete"])
	}
	if len(peers) != 1 {
		t.Errorf("Expected 1 peer; got %d", len(peers))
	}
	if _, ok := peers[0]["peer id"]; ok {
		t.Errorf("Expected no peer id; got %v", peers[0]["peer id"])
	}
	if peers[0]["ip"] != "127.0.0.1" {
		t.Errorf("Expected ip = %v; got %v", "127.0.0.1", peers[0]["ip"])
	}
	if peers[0]["port"] != 1 {
		t.Errorf("Expected port = %v; got %v", 1, peers[0]["port"])
	}

	params.Set("peer_id", "00000000000000000002")
	_, peers, err = sendAnnounce(testNetAddress6, params)
	if err != nil {
		t.Fatal(err)
	}

	var ips []string
	for _, peer := range peers {
		if peer["port"] != 1 {
			t.Errorf("Expected port = %v; got %v", 1, peer["port"])
		}
		ips = append(ips, peer["ip"].(string))
	}

	sort.Strings(ips)
	if ips[0] != "127.0.0.1" {
		t.Errorf("Expected ip = %v; got %v", "127.0.0.1", ips[0])
	}
	if ips[1] != "::1" {
		t.Errorf("Expected ip = %v; got %v", "::1", ips[1])
	}
}
