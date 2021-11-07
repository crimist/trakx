package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

type Error struct {
	Action        int32
	TransactionID int32
	ErrorString   []uint8
}

func (e *Error) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(8 + len(e.ErrorString))

	if err := binary.Write(&buff, binary.BigEndian, e.Action); err != nil {
		return nil, err
	}
	if err := binary.Write(&buff, binary.BigEndian, e.TransactionID); err != nil {
		return nil, err
	}
	if err := binary.Write(&buff, binary.BigEndian, e.ErrorString); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (e *Error) Unmarshall(data []byte, size int) error {
	e.ErrorString = make([]uint8, (size - 8))
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &e.Action); err != nil {
		return errors.Wrap(err, "failed to decode error action")
	}
	if err := binary.Read(reader, binary.BigEndian, &e.TransactionID); err != nil {
		return errors.Wrap(err, "failed to decode error transaction id")
	}
	if err := binary.Read(reader, binary.BigEndian, &e.ErrorString); err != nil {
		return errors.Wrap(err, "failed to decode error string")
	}

	return nil
}
