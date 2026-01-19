package udp

import (
	"net"

	"github.com/crimist/trakx/internal/tracker/udp/udpprotocol"
	"go.uber.org/zap"
)

func (tracker *Tracker) error(remote *net.UDPAddr, message []byte, TransactionID int32) {
	tracker.collector.ErrorResponse()

	protoError := udpprotocol.ErrorResponse{
		Action:        udpprotocol.ActionError,
		TransactionID: TransactionID,
		ErrorString:   message,
	}

	data, err := protoError.Marshal()
	if err != nil {
		zap.L().Error("failed to marshal error packet", zap.Error(err))
		tracker.socket.WriteToUDP([]byte("catastrophic failure"), remote)
	} else {
		tracker.socket.WriteToUDP(data, remote)
	}
}
