package udptracker

import (
	"fmt"
	"net"
	"testing"

	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"github.com/davecgh/go-spew/spew"
)

// TODO: resume here, writing tests for UDP tracker

func TestConnect(t *testing.T) {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", testAddress, testPort))
	if err != nil {
		t.Fatal("Error resolving UDP address:", err.Error())
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatal("Error connecting to UDP server", err.Error())
	}
	defer conn.Close()

	connectReq := udpprotocol.ConnectRequest{
		ProtcolID:     udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 0x1337,
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
		t.Error("Expected action = 0; got", connectResp.Action)
	}
	if connectResp.TransactionID != 0x1337 {
		t.Error("Expected transaction ID = 0x1337; got", connectResp.TransactionID)
	}

	spew.Dump(connectResp)
}
