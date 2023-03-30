package udptracker

import (
	"net"

	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/udptracker/protocol"
	"go.uber.org/zap"
)

func (tracker *Tracker) sendError(remote *net.UDPAddr, message string, TransactionID int32) {
	stats.ServerErrors.Add(1)

	protoError := protocol.Error{
		Action:        protocol.ActionError,
		TransactionID: TransactionID,
		ErrorString:   []byte("internal error"),
	}

	data, err := protoError.Marshall()
	if err != nil {
		zap.L().DPanic("failed to marshal error packet", zap.Error(err))
		tracker.socket.WriteToUDP([]byte("catastrophic failure"), remote)
	} else {
		tracker.socket.WriteToUDP(data, remote)
	}
}
