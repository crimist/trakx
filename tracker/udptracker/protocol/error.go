package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

// BitTorrent UDP tracker server error
type Error struct {
	Action        Action
	TransactionID int32
	ErrorString   []uint8
}

// Marshall encodes an Error to a byte slice.
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

// Unmarshall decodes a byte slice into an Error.
func (e *Error) Unmarshall(data []byte) error {
	e.ErrorString = make([]uint8, (len(data) - 8))
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
