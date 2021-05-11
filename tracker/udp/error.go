package udp

import (
	"bytes"
	"encoding/binary"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

type udperror struct {
	Action        int32
	TransactionID int32
	ErrorString   []uint8
}

func (e *udperror) marshall() ([]byte, error) {
	buff := new(bytes.Buffer)
	buff.Grow(8 + len(e.ErrorString))

	if err := binary.Write(buff, binary.BigEndian, e.Action); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, e.TransactionID); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, e.ErrorString); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

type cerrFields map[string]interface{}

func (u *UDPTracker) newClientError(msg string, TransactionID int32, fieldMap ...cerrFields) []byte {
	storage.Expvar.ClientErrors.Add(1)

	if config.Conf.LogLevel.Debug() {
		fields := []zap.Field{zap.String("msg", msg)}
		if len(fieldMap) == 1 {
			for k, v := range fieldMap[0] {
				fields = append(fields, zap.Any(k, v))
			}
		}

		config.Logger.Info("Client Err", fields...)
	}

	e := udperror{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte(msg),
	}

	data, err := e.marshall()
	if err != nil {
		config.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}

func (u *UDPTracker) newServerError(msg string, err error, TransactionID int32) []byte {
	storage.Expvar.Errors.Add(1)

	e := udperror{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte("internal err"),
	}
	config.Logger.Error(msg, zap.Error(err))

	data, err := e.marshall()
	if err != nil {
		config.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}
