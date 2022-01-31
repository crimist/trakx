package tracker

import (
	"bytes"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/udp/protocol"
	"github.com/go-torrent/bencode"
)

const (
	udptimeout    = 200 * time.Millisecond
	udptimeoutmsg = "UDP tracker not running"
)

var client = &http.Client{
	Timeout: 1 * time.Second,
}

// sets up traker for next tests
func TestRunTracker(t *testing.T) {
	intMax := int(math.Pow(2, 32))
	int64Max := int64(math.Pow(2, 32))

	// set config
	config.Conf.LogLevel = "debug"

	config.Conf.Debug.PprofPort = 0
	config.Conf.Debug.ExpvarInterval = 0
	config.Conf.Debug.NofileLimit = 0
	config.Conf.Debug.PeerChanInit = 0
	config.Conf.Debug.CheckConnIDs = true

	config.Conf.Tracker.Announce = 0
	config.Conf.Tracker.AnnounceFuzz = 1
	config.Conf.Tracker.HTTP.Mode = "enabled"
	config.Conf.Tracker.HTTP.Port = 1337
	config.Conf.Tracker.HTTP.ReadTimeout = 10
	config.Conf.Tracker.HTTP.WriteTimeout = 10
	config.Conf.Tracker.HTTP.Threads = 1
	config.Conf.Tracker.UDP.Enabled = true
	config.Conf.Tracker.UDP.Port = 1337
	config.Conf.Tracker.UDP.Threads = 1
	config.Conf.Tracker.Numwant.Default = 100
	config.Conf.Tracker.Numwant.Limit = 100

	config.Conf.Database.Type = "gomap"
	config.Conf.Database.Backup = "none"
	config.Conf.Database.Peer.Trim = intMax
	config.Conf.Database.Peer.Write = 0
	config.Conf.Database.Peer.Timeout = int64Max
	config.Conf.Database.Conn.Trim = intMax
	config.Conf.Database.Conn.Timeout = int64Max

	// run tracker
	t.Log("Running tracker and waiting 1000ms")
	go Run()
	time.Sleep(1000 * time.Millisecond) // wait for run to complete
}

func TestHTTPAnnounce(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:1337/announce", nil)
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

	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Skip("HTTP tracker not running")
		}
		t.Fatal(err)
	}

	// Parse it
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	resp.Body.Close()

	if len(body) == 0 {
		t.Fatal("body empty")
	}

	var decoded map[string]interface{}
	if err = bencode.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	if peer["peer id"] != "QB123456789012345678" {
		t.Error("PeerID should be QB123456789012345678 got", peer["peer id"])
	}
	if peer["ip"] != "127.0.0.1" {
		t.Error("ip should be 127.0.0.1 got", peer["ip"])
	}
	if peer["port"] != 1234 {
		t.Error("port should be 1234 got", peer["port"])
	}
}

func TestHTTPAnnounceCompact(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:1337/announce", nil)
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

	resp, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Skip("HTTP tracker not running")
		}
		t.Fatal(err)
	}

	// Parse it
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var decoded map[string]interface{}
	if err = bencode.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}

	if _, ok := decoded["failure reason"]; ok {
		t.Error("Tracker error:", decoded["failure reason"])
	}

	peerBytes := []byte(decoded["peers"].(string))
	if len(peerBytes) != 6 {
		t.Fatal("len(peers) should be 6 got", len(peerBytes))
	}
	if !bytes.Equal(peerBytes[0:4], []byte{127, 0, 0, 1}) {
		t.Error("ip should be [127, 0, 0, 1] got", peerBytes[4:6])
	}
	if !bytes.Equal(peerBytes[4:6], []byte{0x04, 0xD2}) {
		t.Error("port should be [4, 210] got", peerBytes[4:6])
	}
}

// UDP

func TestUDPAnnounce(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}

	conn.SetWriteDeadline(time.Now().Add(udptimeout))
	conn.SetReadDeadline(time.Now().Add(udptimeout))

	c := protocol.Connect{
		ProtcolID:     0x41727101980,
		Action:        0,
		TransactionID: 1337,
	}

	data, err := c.Marshall()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write(data); err != nil {
		t.Fatal(err)
	}
	_, err = conn.Read(packet)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "connection refused") {
			t.Skip(udptimeoutmsg)
		}
		t.Fatal(err)
	}

	cr := protocol.ConnectResp{}
	cr.Unmarshall(packet)

	if cr.Action == 3 {
		e := protocol.Error{}
		e.Unmarshall(packet)
		t.Error("Tracker err:", string(e.ErrorString))
	}

	if cr.TransactionID != c.TransactionID {
		t.Error("Invalid transactionID should be", c.TransactionID, "but got", cr.TransactionID)
	}
	if cr.Action != 0 {
		t.Error("Invalid action should be 0 but got", cr.Action)
	}

	a := protocol.Announce{
		ConnectionID:  cr.ConnectionID,
		Action:        1,
		TransactionID: 7331,
		InfoHash:      [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
		PeerID:        [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
		Downloaded:    100,
		Left:          100,
		Uploaded:      50,
		Event:         2,
		IP:            0,
		Key:           0xDEADBEEF,
		NumWant:       1,
		Port:          0xAABB,
	}

	data, err = a.Marshall()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write(data); err != nil {
		t.Error(err)
	}
	_, err = conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	ar := protocol.AnnounceResp{}
	ar.Unmarshall(packet)

	if ar.Action == 3 {
		e := protocol.Error{}
		e.Unmarshall(packet)
		t.Error("Tracker err:", string(e.ErrorString))
		return
	}

	if ar.TransactionID != a.TransactionID {
		t.Error("Invalid transactionID should be", a.TransactionID, "but got", ar.TransactionID)
	}
	if ar.Action != 1 {
		t.Error("Invalid action should be 1 but got", ar.Action)
	}
	if ar.Leechers != 1 {
		t.Error("Invalid leechers should be 1 but got", ar.Leechers)
	}
	if ar.Seeders != 0 {
		t.Error("Invalid seeders should be 1 but got", ar.Seeders)
	}

	if len(ar.Peers) < 1 {
		t.Error("No peers")
		return
	}

	if !bytes.Equal(ar.Peers[4:6], []byte{0xAA, 0xBB}) {
		t.Error("Invalid peer port should be 0xAABB but got", ar.Peers[4:6])
	}
	if !bytes.Equal(ar.Peers[0:4], []byte{127, 0, 0, 1}) {
		t.Error("Invalid peer ip should be 127.0.0.1 but got", ar.Peers[0:4])
	}
}

func TestUDPBadAction(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}

	conn.SetWriteDeadline(time.Now().Add(udptimeout))
	conn.SetReadDeadline(time.Now().Add(udptimeout))

	c := protocol.Connect{
		ProtcolID:     0x41727101980,
		Action:        0,
		TransactionID: 1337,
	}

	data, err := c.Marshall()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write(data); err != nil {
		if strings.Contains(err.Error(), "i/o timeout") {
			t.Skip(udptimeoutmsg)
		}
		t.Fatal(err)
	}
	_, err = conn.Read(packet)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "connection refused") {
			t.Skip(udptimeoutmsg)
		}
		t.Fatal(err)
	}

	cr := protocol.ConnectResp{}
	cr.Unmarshall(packet)

	c = protocol.Connect{
		ProtcolID:     cr.ConnectionID,
		Action:        0xBAD,
		TransactionID: 0xDEAD,
	}

	data, err = c.Marshall()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write(data); err != nil {
		t.Error(err)
	}
	s, err := conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	e := protocol.Error{}
	e.Unmarshall(packet[:s])

	if !bytes.Equal(e.ErrorString, []byte("bad action")) {
		t.Error("Tracker err should be 'bad action' but got:", string(e.ErrorString))
	}
}

func TestUDPBadConnID(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}

	conn.SetWriteDeadline(time.Now().Add(udptimeout))
	conn.SetReadDeadline(time.Now().Add(udptimeout))

	a := protocol.Announce{
		ConnectionID:  0xBAD, // bad connid
		Action:        1,
		TransactionID: 0xDEAD,
	}

	data, err := a.Marshall()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write(data); err != nil {
		if strings.Contains(err.Error(), "i/o timeout") {
			t.Skip(udptimeoutmsg)
		}
		t.Fatal(err)
	}
	s, err := conn.Read(packet)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "connection refused") {
			t.Skip(udptimeoutmsg)
		}
		t.Fatal(err)
	}

	e := protocol.Error{}
	e.Unmarshall(packet[:s])

	if !bytes.Equal(e.ErrorString, []byte("bad connid")) {
		t.Error("Tracker err should be 'bad connid' but got:", string(e.ErrorString))
	}
}

func TestUDPBadPort(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}

	conn.SetWriteDeadline(time.Now().Add(udptimeout))
	conn.SetReadDeadline(time.Now().Add(udptimeout))

	c := protocol.Connect{
		ProtcolID:     0x41727101980,
		Action:        0,
		TransactionID: 1337,
	}

	data, err := c.Marshall()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write(data); err != nil {
		t.Fatal(err)
	}
	_, err = conn.Read(packet)
	if err != nil {
		if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "connection refused") {
			t.Skip(udptimeoutmsg)
		}
		t.Fatal(err)
	}

	cr := protocol.ConnectResp{}
	cr.Unmarshall(packet)

	a := protocol.Announce{
		ConnectionID:  cr.ConnectionID,
		Action:        1,
		TransactionID: 7331,
		InfoHash:      [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
		PeerID:        [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
		Downloaded:    100,
		Left:          100,
		Uploaded:      50,
		Event:         2,
		IP:            0,
		Key:           0xDEADBEEF,
		NumWant:       1,
		Port:          0,
	}

	data, err = a.Marshall()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Write(data); err != nil {
		t.Fatal(err)
	}
	s, err := conn.Read(packet)
	if err != nil {
		t.Fatal(err)
	}

	e := protocol.Error{}
	e.Unmarshall(packet[:s])

	if !bytes.Equal(e.ErrorString, []byte("bad port")) {
		t.Error("Tracker err should be 'bad port' but got:", string(e.ErrorString))
	}
}

func TestUDPTransactionID(t *testing.T) {
	packet := make([]byte, 0xFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}

	conn.SetWriteDeadline(time.Now().Add(udptimeout))
	conn.SetReadDeadline(time.Now().Add(udptimeout))

	c := protocol.Connect{
		ProtcolID:     0x41727101980,
		Action:        0,
		TransactionID: 0xBAD,
	}
	data, err := c.Marshall()
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {

		if _, err = conn.Write(data); err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				t.Skip(udptimeoutmsg)
			}
			t.Fatal(err)
		}
		size, err := conn.Read(packet)
		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") || strings.Contains(err.Error(), "connection refused") {
				t.Skip(udptimeoutmsg)
			}
			t.Fatal(err)
		}

		if size != 16 {
			e := protocol.Error{}
			e.Unmarshall(packet)
			t.Error(i, "Tracker err:", string(e.ErrorString))
		}

		cr := protocol.ConnectResp{}
		cr.Unmarshall(packet)

		if cr.TransactionID != 0xBAD {
			t.Error(i, "Tracker err: tid should be", 0xBAD, "but got", cr.TransactionID)
		}
	}
}
