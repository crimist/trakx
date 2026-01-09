package udp

import (
	"net"
	"testing"

	"github.com/crimist/trakx/tracker/udp/udpprotocol"
)

func connect(t *testing.T, conn *net.UDPConn, connectReq udpprotocol.ConnectRequest) *udpprotocol.ConnectResponse {
	data, err := connectReq.Marshal()
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

	return connectResp
}

func TestConnectSuccess(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectReq := udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	}
	connectResp := connect(t, conn, connectReq)

	if connectResp.Action != udpprotocol.ActionConnect {
		t.Errorf("Expected action = %v; got %v", udpprotocol.ActionConnect, connectResp.Action)
	}
	if connectResp.TransactionID != connectReq.TransactionID {
		t.Errorf("Expected transaction ID = %v; got %v", connectReq.TransactionID, connectResp.TransactionID)
	}
}
