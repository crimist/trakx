package udpprotocol

import (
	"bytes"
	"encoding/binary"

	"github.com/pkg/errors"
)

// UDP tracker error response
type ErrorResponse struct {
	Action        Action
	TransactionID int32
	ErrorString   []uint8
}

// Marshall encodes an ErrorResponse to a byte slice.
func (errResp *ErrorResponse) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(8 + len(errResp.ErrorString))

	if err := binary.Write(&buff, binary.BigEndian, errResp.Action); err != nil {
		return nil, err
	}
	if err := binary.Write(&buff, binary.BigEndian, errResp.TransactionID); err != nil {
		return nil, err
	}
	if err := binary.Write(&buff, binary.BigEndian, errResp.ErrorString); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

// NewErrorResponse decodes a byte slice to an ErrorResponse.
func NewErrorResponse(data []byte) (*ErrorResponse, error) {
	errResp := &ErrorResponse{}
	errResp.ErrorString = make([]uint8, (len(data) - 8))
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &errResp.Action); err != nil {
		return nil, errors.Wrap(err, "failed to decode error action")
	}
	if err := binary.Read(reader, binary.BigEndian, &errResp.TransactionID); err != nil {
		return nil, errors.Wrap(err, "failed to decode error transaction id")
	}
	if err := binary.Read(reader, binary.BigEndian, &errResp.ErrorString); err != nil {
		return nil, errors.Wrap(err, "failed to decode error string")
	}

	return errResp, nil
}
