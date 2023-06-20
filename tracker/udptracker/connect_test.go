package udptracker

import (
	"testing"

	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"github.com/davecgh/go-spew/spew"
)

func TestConnect(t *testing.T) {
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

	if connectResp.Action != udpprotocol.ActionConnect {
		t.Errorf("Expected action = %v; got %v", udpprotocol.ActionConnect, connectResp.Action)
	}
	if connectResp.TransactionID != transaction {
		t.Errorf("Expected transaction ID = %v; got %v", transaction, connectResp.TransactionID)
	}

	spew.Dump(connectResp)
}
