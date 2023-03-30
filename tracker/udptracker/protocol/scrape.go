package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/crimist/trakx/tracker/storage"
	"github.com/pkg/errors"
)

// BitTorrent UDP tracker announce
type Scrape struct {
	ConnectionID  int64
	Action        Action
	TransactionID int32
	InfoHashes    []storage.Hash
}

// Unmarshall decodes a byte slice into a Scrape.
func (s *Scrape) Unmarshall(data []byte) error {
	s.InfoHashes = make([]storage.Hash, (len(data)-16)/20)
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &s.ConnectionID); err != nil {
		return errors.Wrap(err, "failed to decode scrape connection id")
	}
	if err := binary.Read(reader, binary.BigEndian, &s.Action); err != nil {
		return errors.Wrap(err, "failed to decode scrape action")
	}
	if err := binary.Read(reader, binary.BigEndian, &s.TransactionID); err != nil {
		return errors.Wrap(err, "failed to decode scrape transaction id")
	}
	if err := binary.Read(reader, binary.BigEndian, &s.InfoHashes); err != nil {
		return errors.Wrap(err, "failed to decode scrape infohashes")
	}

	return nil
}

// ScrapeInfo holds the information for each infohash in the scrape response
type ScrapeInfo struct {
	Complete   int32
	Incomplete int32
	Downloaded int32
}

// BitTorrent UDP tracker scrape response
type ScrapeResp struct {
	Action        Action
	TransactionID int32
	Info          []ScrapeInfo
}

// Marshall encodes a ScrapeResp to a byte slice.
func (sr *ScrapeResp) Marshall() ([]byte, error) {
	buff := bytes.NewBuffer(make([]byte, 8+len(sr.Info)*12))

	if err := binary.Write(buff, binary.BigEndian, sr.Action); err != nil {
		return nil, errors.Wrap(err, "failed to encode scrape response action")
	}
	if err := binary.Write(buff, binary.BigEndian, sr.TransactionID); err != nil {
		return nil, errors.Wrap(err, "failed to encode scrape response transaction id")
	}
	if err := binary.Write(buff, binary.BigEndian, sr.Info); err != nil {
		return nil, errors.Wrap(err, "failed to encode scrape response info")
	}

	return buff.Bytes(), nil
}
