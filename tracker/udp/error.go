package udp

import (
	"bytes"
	"encoding/binary"

	"github.com/syc0x00/trakx/tracker/storage"
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
	storage.AddExpval(&storage.Expvar.Clienterrs, 1)

	if !u.conf.Trakx.Prod {
		fields := []zap.Field{zap.String("msg", msg)}
		if len(fieldMap) == 1 {
			for k, v := range fieldMap[0] {
				fields = append(fields, zap.Any(k, v))
			}
		}

		u.logger.Info("Client Err", fields...)
	}

	e := udperror{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte(msg),
	}

	data, err := e.marshall()
	if err != nil {
		u.logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}

func (u *UDPTracker) newServerError(msg string, err error, TransactionID int32) []byte {
	storage.AddExpval(&storage.Expvar.Errs, 1)

	e := udperror{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte("internal err"),
	}
	u.logger.Error(msg, zap.Error(err))

	data, err := e.marshall()
	if err != nil {
		u.logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}
