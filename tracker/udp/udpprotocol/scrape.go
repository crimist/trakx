package udpprotocol

import (
	"bytes"
	"encoding/binary"

	"github.com/crimist/trakx/storage"
	"github.com/pkg/errors"
)

// UDP tracker scrape request
type ScrapeRequest struct {
	ConnectionID  int64
	Action        Action
	TransactionID int32
	InfoHashes    []storage.Hash
}

// NewScrapeRequest creates a new ScrapeRequest from a byte slice.
func NewScrapeRequest(data []byte) (*ScrapeRequest, error) {
	scrapeReq := &ScrapeRequest{}
	scrapeReq.InfoHashes = make([]storage.Hash, (len(data)-16)/20)
	reader := bytes.NewReader(data)

	if err := binary.Read(reader, binary.BigEndian, &scrapeReq.ConnectionID); err != nil {
		return nil, errors.Wrap(err, "failed to decode scrape connection id")
	}
	if err := binary.Read(reader, binary.BigEndian, &scrapeReq.Action); err != nil {
		return nil, errors.Wrap(err, "failed to decode scrape action")
	}
	if err := binary.Read(reader, binary.BigEndian, &scrapeReq.TransactionID); err != nil {
		return nil, errors.Wrap(err, "failed to decode scrape transaction id")
	}
	if err := binary.Read(reader, binary.BigEndian, &scrapeReq.InfoHashes); err != nil {
		return nil, errors.Wrap(err, "failed to decode scrape infohashes")
	}

	return scrapeReq, nil
}

// ScrapeResponseInfo holds the information for each infohash in the scrape response
type ScrapeResponseInfo struct {
	Complete   int32
	Incomplete int32
	Downloaded int32
}

// UDP tracker scrape response
type ScrapeResponse struct {
	Action        Action
	TransactionID int32
	Info          []ScrapeResponseInfo
}

// Marshal encodes a ScrapeResp to a byte slice.
func (sr *ScrapeResponse) Marshal() ([]byte, error) {
	var buff bytes.Buffer
	buff.Grow(8 + len(sr.Info)*12)

	if err := binary.Write(&buff, binary.BigEndian, sr.Action); err != nil {
		return nil, errors.Wrap(err, "failed to encode scrape response action")
	}
	if err := binary.Write(&buff, binary.BigEndian, sr.TransactionID); err != nil {
		return nil, errors.Wrap(err, "failed to encode scrape response transaction id")
	}
	if err := binary.Write(&buff, binary.BigEndian, sr.Info); err != nil {
		return nil, errors.Wrap(err, "failed to encode scrape response info")
	}

	return buff.Bytes(), nil
}
