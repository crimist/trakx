package udpprotocol

import (
	"bytes"
	"encoding/binary"

	"github.com/crimist/trakx/storage"
	"github.com/pkg/errors"
)

// UDP tracker announce request
type AnnounceRequest struct {
	ConnectionID  int64
	Action        Action
	TransactionID int32
	InfoHash      storage.Hash
	PeerID        storage.PeerID
	Downloaded    int64
	Left          int64
	Uploaded      int64
	Event         AnnounceEvent
	IP            uint32
	Key           uint32
	NumWant       int32
	Port          uint16
}

// Marshal encodes an AnnounceRequest to a byte slice.
func (a *AnnounceRequest) Marshal() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(98)
	if err := binary.Write(&buff, binary.BigEndian, a); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce")
	}
	return buff.Bytes(), nil
}

// NewAnnounceRequest decodes a byte slice into an AnnounceRequest.
func NewAnnounceRequest(data []byte) (*AnnounceRequest, error) {
	announce := &AnnounceRequest{}
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, announce); err != nil {
		errors.Wrap(err, "failed to decode announce")
	}

	return announce, nil
}

// UDP tracker announce response
type AnnounceResponse struct {
	Action        Action
	TransactionID int32
	Interval      int32
	Leeches       int32
	Seeds         int32
	Peers         []byte
}

// Marshal encodes an AnnounceResp to a byte slice.
func (ar *AnnounceResponse) Marshal() ([]byte, error) {
	var buff bytes.Buffer

	if err := binary.Write(&buff, binary.BigEndian, ar.Action); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response action")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.TransactionID); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response transaction id")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Interval); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response interval")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Leeches); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response leeches")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Seeds); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response seeders")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Peers); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response peers")
	}

	return buff.Bytes(), nil
}

// NewAnnounceResponse decodes a byte slice into an AnnounceResponse.
func NewAnnounceResponse(data []byte) (*AnnounceResponse, error) {
	announceResp := &AnnounceResponse{}
	announceResp.Peers = make([]byte, (len(data) - 20))
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &announceResp.Action); err != nil {
		return nil, errors.Wrap(err, "failed to decode announce response action")
	}
	if err := binary.Read(reader, binary.BigEndian, &announceResp.TransactionID); err != nil {
		return nil, errors.Wrap(err, "failed to decode announce response transaction id")
	}
	if err := binary.Read(reader, binary.BigEndian, &announceResp.Interval); err != nil {
		return nil, errors.Wrap(err, "failed to decode announce response interval")
	}
	if err := binary.Read(reader, binary.BigEndian, &announceResp.Leeches); err != nil {
		return nil, errors.Wrap(err, "failed to decode announce response leeches")
	}
	if err := binary.Read(reader, binary.BigEndian, &announceResp.Seeds); err != nil {
		return nil, errors.Wrap(err, "failed to decode announce response seeders")
	}
	if err := binary.Read(reader, binary.BigEndian, &announceResp.Peers); err != nil {
		return nil, errors.Wrap(err, "failed to decode announce response peers")
	}

	return announceResp, nil
}
