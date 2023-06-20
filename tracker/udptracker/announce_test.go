package udptracker

import (
	"testing"

	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"github.com/davecgh/go-spew/spew"
)

// TODO: resume here, writing tests for UDP tracker

func TestAnnounce(t *testing.T) {
	var transaction int32 = 1

	conn, err := dialTestTracker()
	if err != nil {
		t.Fatal("Error connecting to test UDP tracker", err.Error())
	}
	defer conn.Close()

	connectReq := udpprotocol.ConnectRequest{
		ProtcolID:     udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: transaction,
	}
	data, err := connectReq.Marshall()
	if err != nil {
		t.Fatal("Error marshalling connect request:", err.Error())
	}
	_, err = conn.Write(data)
	if err != nil {
		t.Fatal("Error sending message to UDP server", err.Error())
	}

	data = make([]byte, 1024)
	conn.Read(data)
	connectResp, err := udpprotocol.NewConnectResponse(data)
	if err != nil {
		t.Fatal("Error unmarshalling connect response:", err.Error())
	}
	transaction++

	announceReq := udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: transaction,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          4096,
	}
	data, err = announceReq.Marshall()
	if err != nil {
		t.Fatal("Error marshalling connect request:", err.Error())
	}
	_, err = conn.Write(data)
	if err != nil {
		t.Fatal("Error sending message to UDP server", err.Error())
	}

	data = make([]byte, 1024)
	conn.Read(data)
	announceResp, err := udpprotocol.NewAnnounceResponse(data)
	if err != nil {
		t.Fatal("Error unmarshalling connect response:", err.Error())
	}

	if announceResp.Action != udpprotocol.ActionAnnounce {
		t.Errorf("Expected action = %v; got %v", udpprotocol.ActionAnnounce, announceResp.Action)
	}
	if announceResp.TransactionID != transaction {
		t.Errorf("Expected action = %v; got %v", transaction, announceResp.Action)
	}
	// TransactionID int32
	// Interval      int32
	// Leechers      int32
	// Seeders       int32
	// Peers         []byte
	transaction++

	// announce complete
	// ...

	// announce stopped
	// ...

	spew.Dump(connectResp)
}
