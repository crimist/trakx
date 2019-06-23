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
	if err := binary.Write(buff, binary.BigEndian, e); err != nil {
		return nil, err
	}
	// binary.Write(buff, binary.BigEndian, e.Action)
	// binary.Write(buff, binary.BigEndian, e.TransactionID)
	// binary.Write(buff, binary.BigEndian, e.ErrorString)
	return buff.Bytes(), nil
}

func newClientError(msg string, TransactionID int32) []byte {
	e := Error{
		Action:        3,
		TransactionID: TransactionID,
		ErrorString:   []byte(msg),
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
		ErrorString:   []byte("internal server error"),
	}
	shared.Logger.Error(msg, zap.Error(err))

	data, err := e.Marshall()
	if err != nil {
		shared.Logger.Error("e.Marshall()", zap.Error(err))
	}
	return data
}
