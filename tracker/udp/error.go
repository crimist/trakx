package udp

import (
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/udp/protocol"
	"go.uber.org/zap"
)

type cerrFields map[string]interface{}

func (u *UDPTracker) newClientError(msg string, TransactionID int32, fieldMap ...cerrFields) []byte {
	stats.ClientErrors.Add(1)

	if config.Config.LogLevel.Debug() {
		fields := []zap.Field{zap.String("msg", msg)}
		if len(fieldMap) == 1 {
			for k, v := range fieldMap[0] {
				fields = append(fields, zap.Any(k, v))
			}
		}

		config.Logger.Info("Client Err", fields...)
	}

	e := protocol.Error{
		Action:        protocol.ActionError,
		TransactionID: TransactionID,
		ErrorString:   []byte(msg),
	}

	data, err := e.Marshall()
	if err != nil {
		config.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}

func (u *UDPTracker) newServerError(msg string, err error, TransactionID int32) []byte {
	stats.ServerErrors.Add(1)

	e := protocol.Error{
		Action:        protocol.ActionError,
		TransactionID: TransactionID,
		ErrorString:   []byte("internal err"),
	}
	config.Logger.Error(msg, zap.Error(err))

	data, err := e.Marshall()
	if err != nil {
		config.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}
