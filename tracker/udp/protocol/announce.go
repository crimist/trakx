package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/crimist/trakx/tracker/storage"
	"github.com/pkg/errors"
)

type event int32

const (
	None      event = 0
	Completed event = 1
	Started   event = 2
	Stopped   event = 3
)

type Announce struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
	InfoHash      storage.Hash
	PeerID        storage.PeerID
	Downloaded    int64
	Left          int64
	Uploaded      int64
	Event         event
	IP            uint32
	Key           uint32
	NumWant       int32
	Port          uint16
	// Extensions    uint16
}

func (a *Announce) Marshall() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(98)
	if err := binary.Write(&buff, binary.BigEndian, a); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce")
	}
	return buff.Bytes(), nil
}

func (a *Announce) Unmarshall(data []byte) error {
	if err := binary.Read(bytes.NewReader(data), binary.BigEndian, a); err != nil {
		errors.Wrap(err, "failed to decode announce")
	}

	return nil
}

type AnnounceResp struct {
	Action        int32
	TransactionID int32
	Interval      int32
	Leechers      int32
	Seeders       int32
	Peers         []byte
}

func (ar *AnnounceResp) Marshall() ([]byte, error) {
	var buff bytes.Buffer

	if err := binary.Write(&buff, binary.BigEndian, ar.Action); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response action")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.TransactionID); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response transaction id")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Interval); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response iterval")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Leechers); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response leeches")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Seeders); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response seeders")
	}
	if err := binary.Write(&buff, binary.BigEndian, ar.Peers); err != nil {
		return nil, errors.Wrap(err, "failed to encode announce response peers")
	}

	return buff.Bytes(), nil
}

func (ar *AnnounceResp) Unmarshall(data []byte, size int) error {
	ar.Peers = make([]byte, (size - 20))

	reader := bytes.NewReader(data)
	if err := binary.Read(reader, binary.BigEndian, &ar.Action); err != nil {
		return errors.Wrap(err, "failed to decode announce response action")
	}
	if err := binary.Read(reader, binary.BigEndian, &ar.TransactionID); err != nil {
		return errors.Wrap(err, "failed to decode announce response transaction id")
	}
	if err := binary.Read(reader, binary.BigEndian, &ar.Interval); err != nil {
		return errors.Wrap(err, "failed to decode announce response iterval")
	}
	if err := binary.Read(reader, binary.BigEndian, &ar.Leechers); err != nil {
		return errors.Wrap(err, "failed to decode announce response leeches")
	}
	if err := binary.Read(reader, binary.BigEndian, &ar.Seeders); err != nil {
		return errors.Wrap(err, "failed to decode announce response seeders")
	}
	if err := binary.Read(reader, binary.BigEndian, &ar.Peers); err != nil {
		return errors.Wrap(err, "failed to decode announce response peers")
	}

	return nil
}
