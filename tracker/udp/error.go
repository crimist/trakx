package udp

import (
	"bytes"
	"encoding/binary"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type udperror struct {
	Action        int32
	TransactionID int32
	ErrorString   []uint8
}

func (e *udperror) marshall() ([]byte, error) {
	buff := new(bytes.Buffer)
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

func newClientError(msg string, TransactionID int32, fields ...zap.Field) []byte {
	shared.ExpvarClienterrs++

	if shared.Env == shared.Dev {
		fields = append(fields, zap.String("msg", msg))
		shared.Logger.Info("Client Err", fields...)
	}

	e := udperror{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte(msg),
	}

	data, err := e.marshall()
	if err != nil {
		shared.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}

func newServerError(msg string, err error, TransactionID int32) []byte {
	shared.ExpvarErrs++

	e := udperror{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte("internal err"),
	}
	shared.Logger.Error(msg, zap.Error(err))

	data, err := e.marshall()
	if err != nil {
		shared.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}
