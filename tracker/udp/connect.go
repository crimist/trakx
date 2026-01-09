package udp

import (
	"net"
	"net/netip"

	"github.com/crimist/trakx/tracker/udp/udpprotocol"
	"go.uber.org/zap"
)

func (tracker *Tracker) connect(udpAddr *net.UDPAddr, addrPort netip.AddrPort, transactionID int32, data []byte) {
	tracker.collector.Connect()

	connectRequest, err := udpprotocol.NewConnectRequest(data)
	if err != nil {
		tracker.error(udpAddr, []byte("failed to parse connect request"), transactionID)
		zap.L().Debug("client sent invalid connect request", zap.Binary("packet", data), zap.Error(err), zap.Any("remote", addrPort))
		return
	}

	connectionID := tracker.connections.Create(addrPort)

	resp := udpprotocol.ConnectResponse{
		Action:        udpprotocol.ActionConnect,
		TransactionID: connectRequest.TransactionID,
		ConnectionID:  connectionID,
	}

	marshalledResp, err := resp.Marshal()
	if err != nil {
		tracker.error(udpAddr, []byte("failed to marshall connect response"), connectRequest.TransactionID)
		zap.L().Error("failed to marshall connect response", zap.Error(err), zap.Any("connect", connectRequest), zap.Any("remote", udpAddr))
		return
	}

	tracker.socket.WriteToUDP(marshalledResp, udpAddr)
}
