package udptracker

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"go.uber.org/zap"
)

func (tracker *Tracker) connect(udpAddr *net.UDPAddr, addrPort netip.AddrPort, transactionID int32, data []byte) {
	if tracker.stats != nil {
		tracker.stats.Connects.Add(1)
	}

	connectRequest, err := udpprotocol.NewConnectRequest(data)
	if err != nil {
		tracker.sendError(udpAddr, "failed to parse connect request", transactionID)
		zap.L().Debug("client sent invalid connect request", zap.Binary("packet", data), zap.Error(err), zap.Any("remote", addrPort))
		return
	}

	connectionID := rand.Int63()
	tracker.connCache.Set(connectionID, addrPort)

	resp := udpprotocol.ConnectResponse{
		Action:        udpprotocol.ActionConnect,
		TransactionID: connectRequest.TransactionID,
		ConnectionID:  connectionID,
	}

	marshalledResp, err := resp.Marshall()
	if err != nil {
		tracker.sendError(udpAddr, "failed to marshall connect response", connectRequest.TransactionID)
		zap.L().Error("failed to marshall connect response", zap.Error(err), zap.Any("connect", connectRequest), zap.Any("remote", udpAddr))
		return
	}

	tracker.socket.WriteToUDP(marshalledResp, udpAddr)
}
