package tracker_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/go-torrent/bencode"
	"github.com/crimist/trakx/utils"
)

var client = &http.Client{
	Timeout: 2 * time.Second,
}

func TestHTTPAnnounce(t *testing.T) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:1337/announce", nil)
	if err != nil {
		t.Error(err)
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
		t.Error(err)
	}

	// Parse it
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	resp.Body.Close()

	if len(body) == 0 {
		t.Error("body empty")
	}

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
	if bytes.Compare(peerBytes[0:4], []byte{127, 0, 0, 1}) != 0 {
		t.Error("ip should be [127, 0, 0, 1] got", peerBytes[4:6])
	}
	if bytes.Compare(peerBytes[4:6], []byte{0x04, 0xD2}) != 0 {
		t.Error("port should be [4, 210] got", peerBytes[4:6])
	}
}

// UDP

type Error struct {
	X struct {
		Action        int32
		TransactionID int32
	}
	ErrorString []uint8
}

func (e *Error) Unmarshall(data []byte, size int) {
	e.ErrorString = make([]uint8, (size - 8))
	readerX := bytes.NewReader(data[:8])
	err := binary.Read(readerX, binary.BigEndian, &e.X)
	if err != nil {
		panic(err)
	}

	readerErrorString := bytes.NewReader(data[8:])
	err = binary.Read(readerErrorString, binary.BigEndian, &e.ErrorString)
	if err != nil {
		panic(err)
	}
}

type Connect struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
}

func (c *Connect) Marshall() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, c.ConnectionID)
	binary.Write(buff, binary.BigEndian, c.Action)
	binary.Write(buff, binary.BigEndian, c.TransactionID)
	return buff.Bytes()
}

type ConnectResp struct {
	Action        int32
	TransactionID int32
	ConnectionID  int64
}

func (cr *ConnectResp) Unmarshall(data []byte) {
	reader := bytes.NewReader(data)
	err := binary.Read(reader, binary.BigEndian, cr)
	if err != nil {
		panic(err)
	}
}

type Announce struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
	InfoHash      [20]byte
	PeerID        [20]byte
	Downloaded    int64
	Left          int64
	Uploaded      int64
	Event         int32
	IP            uint32
	Key           uint32
	NumWant       int32
	Port          uint16
	Extensions    uint16
}

func (a *Announce) Marshall() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, a.ConnectionID)
	binary.Write(buff, binary.BigEndian, a.Action)
	binary.Write(buff, binary.BigEndian, a.TransactionID)
	binary.Write(buff, binary.BigEndian, a.InfoHash)
	binary.Write(buff, binary.BigEndian, a.PeerID)
	binary.Write(buff, binary.BigEndian, a.Downloaded)
	binary.Write(buff, binary.BigEndian, a.Left)
	binary.Write(buff, binary.BigEndian, a.Uploaded)
	binary.Write(buff, binary.BigEndian, a.Event)
	binary.Write(buff, binary.BigEndian, a.IP)
	binary.Write(buff, binary.BigEndian, a.Key)
	binary.Write(buff, binary.BigEndian, a.NumWant)
	binary.Write(buff, binary.BigEndian, a.Port)
	binary.Write(buff, binary.BigEndian, a.Extensions)
	return buff.Bytes()
}

type Peer struct {
	IP   uint32
	Port uint16
}

type AnnounceResp struct {
	X struct {
		Action        int32
		TransactionID int32
		Interval      int32
		Leechers      int32
		Seeders       int32
	}
	Peers []Peer
}

func (ar *AnnounceResp) Unmarshall(data []byte, size int) {
	ar.Peers = make([]Peer, (size-20)/6)
	readerX := bytes.NewReader(data[:20])
	err := binary.Read(readerX, binary.BigEndian, &ar.X)
	if err != nil {
		panic(err)
	}

	readerPeers := bytes.NewReader(data[20:])
	err = binary.Read(readerPeers, binary.BigEndian, &ar.Peers)
	if err != nil {
		panic(err)
	}
}

func TestUDPAnnounce(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Error(err)
	}

	c := Connect{
		ConnectionID:  0x41727101980,
		Action:        0,
		TransactionID: 1337,
	}

	if _, err = conn.Write(c.Marshall()); err != nil {
		t.Error(err)
	}
	s, err := conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	cr := ConnectResp{}
	cr.Unmarshall(packet)

	if cr.Action == 3 {
		e := Error{}
		e.Unmarshall(packet, s)
		t.Error("Tracker err:", string(e.ErrorString))
	}

	if cr.TransactionID != c.TransactionID {
		t.Error("Invalid transactionID should be", c.TransactionID, "but got", cr.TransactionID)
	}
	if cr.Action != 0 {
		t.Error("Invalid action should be 0 but got", cr.Action)
	}

	a := Announce{
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
		Port:          1337,
		Extensions:    0,
	}

	if _, err = conn.Write(a.Marshall()); err != nil {
		t.Error(err)
	}
	s, err = conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	ar := AnnounceResp{}
	ar.Unmarshall(packet, s)

	if ar.X.Action == 3 {
		e := Error{}
		e.Unmarshall(packet, s)
		t.Error("Tracker err:", string(e.ErrorString))
		return
	}

	if ar.X.TransactionID != a.TransactionID {
		t.Error("Invalid transactionID should be", a.TransactionID, "but got", ar.X.TransactionID)
	}
	if ar.X.Action != 1 {
		t.Error("Invalid action should be 1 but got", ar.X.Action)
	}
	if ar.X.Leechers != 1 {
		t.Error("Invalid leechers should be 1 but got", ar.X.Leechers)
	}
	if ar.X.Seeders != 0 {
		t.Error("Invalid seeders should be 1 but got", ar.X.Seeders)
	}

	if len(ar.Peers) < 1 {
		t.Error("No peers")
		return
	}

	if ar.Peers[0].Port != 1337 {
		t.Error("Invalid peer port should be 1337 but got", ar.Peers[0].Port)
	}
	ip := utils.IntToIP(ar.Peers[0].IP)
	ipstr := fmt.Sprintf("%d.%d.%d.%d", ar.Peers[0].IP>>24, uint16(ar.Peers[0].IP)>>16, uint16(ar.Peers[0].IP)>>8, byte(ar.Peers[0].IP))
	if bytes.Compare(ip, net.IPv4(127, 0, 0, 1)) == 0 {
		t.Error("Invalid peer ip should be 127.0.0.1 but got", ip)
	}
	if ipstr != "127.0.0.1" {
		t.Error("Invalid peer ip should be 127.0.0.1 but got", ipstr)
	}
}

func TestUDPBadAction(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Error(err)
	}

	c := Connect{
		ConnectionID:  0x41727101980,
		Action:        0,
		TransactionID: 1337,
	}

	if _, err = conn.Write(c.Marshall()); err != nil {
		t.Error(err)
	}
	_, err = conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	cr := ConnectResp{}
	cr.Unmarshall(packet)

	c = Connect{
		ConnectionID:  cr.ConnectionID,
		Action:        0xBAD,
		TransactionID: 0xDEAD,
	}

	if _, err = conn.Write(c.Marshall()); err != nil {
		t.Error(err)
	}
	s, err := conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	e := Error{}
	e.Unmarshall(packet, s)

	if bytes.Compare(e.ErrorString, []byte("bad action")) != 0 {
		t.Error("Tracker err should be 'bad action' but got:", string(e.ErrorString))
	}
}

func TestUDPBadConnID(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Error(err)
	}

	a := Announce{
		ConnectionID:  0xBAD, // bad connid
		Action:        1,
		TransactionID: 0xDEAD,
	}

	if _, err = conn.Write(a.Marshall()); err != nil {
		t.Error(err)
	}
	s, err := conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	e := Error{}
	e.Unmarshall(packet, s)

	if bytes.Compare(e.ErrorString, []byte("bad connid")) != 0 {
		t.Error("Tracker err should be 'bad connid' but got:", string(e.ErrorString))
	}
}

func TestUDPBadPort(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Error(err)
	}

	c := Connect{
		ConnectionID:  0x41727101980,
		Action:        0,
		TransactionID: 1337,
	}

	if _, err = conn.Write(c.Marshall()); err != nil {
		t.Error(err)
	}
	_, err = conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	cr := ConnectResp{}
	cr.Unmarshall(packet)

	a := Announce{
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
		Extensions:    0,
	}

	if _, err = conn.Write(a.Marshall()); err != nil {
		t.Error(err)
	}
	s, err := conn.Read(packet)
	if err != nil {
		t.Error(err)
	}

	e := Error{}
	e.Unmarshall(packet, s)

	if bytes.Compare(e.ErrorString, []byte("bad port")) != 0 {
		t.Error("Tracker err should be 'bad port' but got:", string(e.ErrorString))
	}
}

func TestUDPTransactionID(t *testing.T) {
	packet := make([]byte, 0xFF)
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:1337")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Error(err)
	}

	c := Connect{
		ConnectionID:  0x41727101980,
		Action:        0,
		TransactionID: 0xBAD,
	}
	data := c.Marshall()

	for i := 0; i < 1000; i++ {

		if _, err = conn.Write(data); err != nil {
			t.Error(err)
		}
		size, err := conn.Read(packet)
		if err != nil {
			t.Error(err)
		}

		if size != 16 {
			e := Error{}
			e.Unmarshall(packet, size)
			t.Error(i, "Tracker err:", string(e.ErrorString))
		}

		cr := ConnectResp{}
		cr.Unmarshall(packet)

		if cr.TransactionID != 0xBAD {
			t.Error(i, "Tracker err: tid should be", 0xBAD, "but got", cr.TransactionID)
		}
	}
}
