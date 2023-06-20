package udpprotocol

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

const (
	ProtocolMagic = 0x41727101980
)

// UDP tracker connect request
type ConnectRequest struct {
	ProtcolID     int64
	Action        Action
	TransactionID int32
}

// Marshall encodes a ConnectRequest to a byte slice.
func (c *ConnectRequest) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(16)
	if err := binary.Write(&buff, binary.BigEndian, c); err != nil {
		return nil, errors.Wrap(err, "failed to encode connect")
	}
	return buff.Bytes(), nil
}

// NewConnectRequest decodes a byte slice into a ConnectRequest.
func NewConnectRequest(data []byte) (*ConnectRequest, error) {
	connect := &ConnectRequest{}
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, connect); err != nil {
		errors.Wrap(err, "failed to decode connect")
	}

	return connect, nil
}

// UDP tracker connect response
type ConnectResponse struct {
	Action        Action
	TransactionID int32
	ConnectionID  int64
}

// Marshall encodes a ConnectResponse to a byte slice.
func (cr *ConnectResponse) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(16)
	if err := binary.Write(&buff, binary.BigEndian, cr); err != nil {
		return nil, errors.Wrap(err, "failed to encode connect")
	}
	return buff.Bytes(), nil
}

// NewConnectResponse decodes a byte slice into a ConnectResponse.
func NewConnectResponse(data []byte) (*ConnectResponse, error) {
	connectResp := &ConnectResponse{}
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, connectResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode connect response")
	}

	return connectResp, nil
}
