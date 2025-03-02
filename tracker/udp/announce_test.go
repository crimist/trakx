package udp

import (
	"bytes"
	"net"
	"testing"

	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker/udp/udpprotocol"
)

func announceSuccess(t *testing.T, conn *net.UDPConn, announceReq udpprotocol.AnnounceRequest) udpprotocol.AnnounceResponse {
	data, err := announceReq.Marshal()
	if err != nil {
		t.Fatal("Error marshalling connect request:", err.Error())
	}
	_, err = conn.Write(data)
	if err != nil {
		t.Fatal("Error sending message to UDP server", err.Error())
	}

	data = make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		t.Fatal("Error reading from UDP server", err.Error())
	}
	data = data[:n]

	announceResp, err := udpprotocol.NewAnnounceResponse(data)
	if err != nil {
		t.Fatal("Error unmarshalling connect response:", err.Error())
	}

	if announceResp.Action != udpprotocol.ActionAnnounce {
		t.Errorf("Expected action = %v; got %v", udpprotocol.ActionAnnounce, announceResp.Action)
	}
	if announceResp.TransactionID != announceReq.TransactionID {
		t.Errorf("Expected action = %v; got %v", announceReq.TransactionID, announceResp.Action)
	}
	if uint(announceResp.Interval) != testTrackerConfig.Interval {
		t.Errorf("Expected interval = %v; got %v", testTrackerConfig.Interval, announceResp.Interval)
	}

	return *announceResp
}

func announceError(t *testing.T, conn *net.UDPConn, announceReq udpprotocol.AnnounceRequest) udpprotocol.ErrorResponse {
	data, err := announceReq.Marshal()
	if err != nil {
		t.Fatal("Error marshalling connect request:", err.Error())
	}
	_, err = conn.Write(data)
	if err != nil {
		t.Fatal("Error sending message to UDP server", err.Error())
	}

	data = make([]byte, 1024)
	n, err := conn.Read(data)
	if err != nil {
		t.Fatal("Error reading from UDP server", err.Error())
	}
	data = data[:n]

	errorResp, err := udpprotocol.NewErrorResponse(data)
	if err != nil {
		t.Fatal("Error unmarshalling connect response:", err.Error())
	}

	return *errorResp
}

func TestAnnounceStarted(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectResp := connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	announceResp := announceSuccess(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0xAABB,
	})

	if announceResp.Leeches != 1 {
		t.Errorf("Expected leeches = %v; got %v", 1, announceResp.Leeches)
	}
	if announceResp.Seeds != 0 {
		t.Errorf("Expected seeds = %v; got %v", 0, announceResp.Seeds)
	}
	if len(announceResp.Peers) != 6 {
		t.Errorf("Expected len(peers) = %v; got %v", 6, len(announceResp.Peers))
	}
	if !bytes.Equal(announceResp.Peers[4:6], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[4:6])
	}
	if !bytes.Equal(announceResp.Peers[0:4], []byte{127, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{127, 0, 0, 1}, announceResp.Peers[0:4])
	}
}

func TestAnnounceStarted6(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress6)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectResp := connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	announceResp := announceSuccess(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0xAABB,
	})

	if announceResp.Leeches != 1 {
		t.Errorf("Expected leeches = %v; got %v", 1, announceResp.Leeches)
	}
	if announceResp.Seeds != 0 {
		t.Errorf("Expected seeds = %v; got %v", 0, announceResp.Seeds)
	}
	if len(announceResp.Peers) != 18 {
		t.Errorf("Expected len(peers) = %v; got %v", 18, len(announceResp.Peers))
	}
	if !bytes.Equal(announceResp.Peers[16:18], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[16:18])
	}
	if !bytes.Equal(announceResp.Peers[0:16], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, announceResp.Peers[0:16])
	}
}

// Test an announce with event = completed
func TestAnnounceCompleteEvent(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectResp := connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	announceResp := announceSuccess(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventCompleted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0xAABB,
	})

	if announceResp.Leeches != 0 {
		t.Errorf("Expected leeches = %v; got %v", 0, announceResp.Leeches)
	}
	if announceResp.Seeds != 1 {
		t.Errorf("Expected seeds = %v; got %v", 1, announceResp.Seeds)
	}
	if len(announceResp.Peers) != 6 {
		t.Errorf("Expected len(peers) = %v; got %v", 6, len(announceResp.Peers))
	}
	if !bytes.Equal(announceResp.Peers[4:6], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[4:6])
	}
	if !bytes.Equal(announceResp.Peers[0:4], []byte{127, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{127, 0, 0, 1}, announceResp.Peers[0:4])
	}
}

// Test an announce where left = 0
func TestAnnounceCompleteLeft(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectResp := connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	announceResp := announceSuccess(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          0,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0xAABB,
	})

	if announceResp.Leeches != 0 {
		t.Errorf("Expected leeches = %v; got %v", 0, announceResp.Leeches)
	}
	if announceResp.Seeds != 1 {
		t.Errorf("Expected seeds = %v; got %v", 1, announceResp.Seeds)
	}
	if len(announceResp.Peers) != 6 {
		t.Errorf("Expected len(peers) = %v; got %v", 6, len(announceResp.Peers))
	}
	if !bytes.Equal(announceResp.Peers[4:6], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[4:6])
	}
	if !bytes.Equal(announceResp.Peers[0:4], []byte{127, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{127, 0, 0, 1}, announceResp.Peers[0:4])
	}
}

func TestAnnounceStopped(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectResp := connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	announceResp := announceSuccess(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          0,
		Uploaded:      1000,
		Event:         udpprotocol.EventStopped,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0xAABB,
	})

	if announceResp.Leeches != 0 {
		t.Errorf("Expected leeches = %v; got %v", 0, announceResp.Leeches)
	}
	if announceResp.Seeds != 0 {
		t.Errorf("Expected seeds = %v; got %v", 0, announceResp.Seeds)
	}
	if len(announceResp.Peers) != 0 {
		t.Errorf("Expected len(peers) = %v; got %v", 0, len(announceResp.Peers))
	}
}

func TestAnnounceInvalidPort(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectResp := connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	errorResp := announceError(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0,
	})

	if !bytes.Equal(errorResp.ErrorString, []byte(fatalInvalidPort)) {
		t.Errorf("Expected error = %v; got %v", fatalInvalidPort, errorResp.ErrorString)
	}
}

// TODO: add test case for multiple peers
