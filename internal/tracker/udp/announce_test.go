package udp

import (
	"bytes"
	"net"
	"testing"

	"github.com/crimist/trakx/internal/storage"
	"github.com/crimist/trakx/internal/tracker/udp/udpprotocol"
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

func TestAnnounce4(t *testing.T) {
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
		InfoHash:      storage.Hash{1},
		PeerID:        storage.PeerID{1},
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

	announceResp = announceSuccess(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{1},
		PeerID:        storage.PeerID{2},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0xAABB,
	})

	if announceResp.Leeches != 2 {
		t.Errorf("Expected leeches = %v; got %v", 2, announceResp.Leeches)
	}
	if announceResp.Seeds != 0 {
		t.Errorf("Expected seeds = %v; got %v", 0, announceResp.Seeds)
	}
	if len(announceResp.Peers) != 12 {
		t.Errorf("Expected len(peers) = %v; got %v", 12, len(announceResp.Peers))
	}
	if !bytes.Equal(announceResp.Peers[4:6], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[4:6])
	}
	if !bytes.Equal(announceResp.Peers[0:4], []byte{127, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{127, 0, 0, 1}, announceResp.Peers[0:4])
	}
	if !bytes.Equal(announceResp.Peers[10:12], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[4:6])
	}
	if !bytes.Equal(announceResp.Peers[6:10], []byte{127, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{127, 0, 0, 1}, announceResp.Peers[0:4])
	}
}

func TestAnnounce6(t *testing.T) {
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
		InfoHash:      storage.Hash{2},
		PeerID:        storage.PeerID{1},
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

	announceResp = announceSuccess(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{2},
		PeerID:        storage.PeerID{2},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          0xAABB,
	})

	if announceResp.Leeches != 2 {
		t.Errorf("Expected leeches = %v; got %v", 2, announceResp.Leeches)
	}
	if announceResp.Seeds != 0 {
		t.Errorf("Expected seeds = %v; got %v", 0, announceResp.Seeds)
	}
	if len(announceResp.Peers) != 36 {
		t.Errorf("Expected len(peers) = %v; got %v", 36, len(announceResp.Peers))
	}
	if !bytes.Equal(announceResp.Peers[16:18], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[16:18])
	}
	if !bytes.Equal(announceResp.Peers[0:16], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, announceResp.Peers[0:16])
	}
	if !bytes.Equal(announceResp.Peers[34:36], []byte{0xAA, 0xBB}) {
		t.Errorf("Expected peer port = %#v; got %#v", []byte{0xAA, 0xBB}, announceResp.Peers[16:18])
	}
	if !bytes.Equal(announceResp.Peers[18:34], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}) {
		t.Errorf("Expected peer ip = %v; got %v", []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, announceResp.Peers[0:16])
	}
}

// Test an announce with event = completed
func TestAnnounceCompleted(t *testing.T) {
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
		InfoHash:      storage.Hash{3},
		PeerID:        storage.PeerID{1},
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
		InfoHash:      storage.Hash{4},
		PeerID:        storage.PeerID{1},
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
		InfoHash:      storage.Hash{5},
		PeerID:        storage.PeerID{1},
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
		InfoHash:      storage.Hash{6},
		PeerID:        storage.PeerID{1},
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
