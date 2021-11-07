package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

type Connect struct {
	ProtcolID     int64 // 0x41727101980
	Action        int32
	TransactionID int32
}

func (c *Connect) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(16)
	if err := binary.Write(&buff, binary.BigEndian, c); err != nil {
		return nil, errors.Wrap(err, "failed to encode connect")
	}
	return buff.Bytes(), nil
}

func (c *Connect) Unmarshall(data []byte) error {
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, c); err != nil {
		return errors.Wrap(err, "failed to decode connect")
	}

	return nil
}

type ConnectResp struct {
	Action        int32
	TransactionID int32
	ConnectionID  int64
}

func (cr *ConnectResp) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(16)
	if err := binary.Write(&buff, binary.BigEndian, cr); err != nil {
		return nil, errors.Wrap(err, "failed to encode connect")
	}
	return buff.Bytes(), nil
}

func (cr *ConnectResp) Unmarshall(data []byte) error {
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, cr); err != nil {
		return errors.Wrap(err, "failed to decode connect")
	}

	return nil
}
