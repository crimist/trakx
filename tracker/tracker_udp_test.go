package tracker

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/crimist/trakx/tracker/udp/protocol"
)

const (
	udptimeout         = 500 * time.Millisecond
	udptimeoutmsg      = "UDP tracker not running"
	announceUDPaddress = "127.0.0.1:1337"
)

func TestUDPAnnounce(t *testing.T) {
	packet := make([]byte, 0xFFFF)
	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
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
		ProtcolID:     protocol.UDPTrackerMagic,
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
	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
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
		ProtcolID:     protocol.UDPTrackerMagic,
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
	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
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
	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
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
		ProtcolID:     protocol.UDPTrackerMagic,
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
	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
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
		ProtcolID:     protocol.UDPTrackerMagic,
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
