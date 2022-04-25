package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

const (
	UDPTrackerMagic = 0x41727101980
)

// BitTorrent UDP tracker connect
type Connect struct {
	ProtcolID     int64
	Action        Action
	TransactionID int32
}

// Marshall encodes a Connect to a byte slice.
func (c *Connect) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(16)
	if err := binary.Write(&buff, binary.BigEndian, c); err != nil {
		return nil, errors.Wrap(err, "failed to encode connect")
	}
	return buff.Bytes(), nil
}

// Unmarshall decodes a byte slice into a Connect.
func (c *Connect) Unmarshall(data []byte) error {
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, c); err != nil {
		return errors.Wrap(err, "failed to decode connect")
	}

	return nil
}

// BitTorrent UDP tracker connect response
type ConnectResp struct {
	Action        Action
	TransactionID int32
	ConnectionID  int64
}

// Marshall encodes a ConnectResp to a byte slice.
func (cr *ConnectResp) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(16)
	if err := binary.Write(&buff, binary.BigEndian, cr); err != nil {
		return nil, errors.Wrap(err, "failed to encode connect")
	}
	return buff.Bytes(), nil
}

// Unmarshall decodes a byte slice into a ConnectResp.
func (cr *ConnectResp) Unmarshall(data []byte) error {
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, cr); err != nil {
		return errors.Wrap(err, "failed to decode connect")
	}

	return nil
}
