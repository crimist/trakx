package udp

import (
	"bytes"
	"encoding/binary"

	"github.com/Syc0x00/Trakx/tracker/shared"
	"go.uber.org/zap"
)

type Error struct {
	Action        int32
	TransactionID int32
	ErrorString   []uint8
}

func (e *Error) Marshall() ([]byte, error) {
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

func newClientError(msg string, TransactionID int32) []byte {
	shared.ExpvarClienterrs++

	e := Error{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte(msg),
	}
	if shared.Env == shared.Dev {
		shared.Logger.Info("Client Err", zap.String("msg", msg))
	}

	data, err := e.Marshall()
	if err != nil {
		shared.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}

func newServerError(msg string, err error, TransactionID int32) []byte {
	shared.ExpvarErrs++

	e := Error{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte("internal err"),
	}
	shared.Logger.Error(msg, zap.Error(err))

	data, err := e.Marshall()
	if err != nil {
		shared.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}
