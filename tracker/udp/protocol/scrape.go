package protocol

import (
	"bytes"
	"encoding/binary"

	"github.com/crimist/trakx/tracker/storage"
)

type Scrape struct {
	Base     Connect
	InfoHash []storage.Hash
}

func (s *Scrape) Unmarshall(data []byte) error {
	if err := binary.Read(bytes.NewReader(data[:16]), binary.BigEndian, &s.Base); err != nil {
		return err
	}

	s.InfoHash = make([]storage.Hash, (len(data)-16)/20)
	return binary.Read(bytes.NewReader(data[16:]), binary.BigEndian, &s.InfoHash)
}

type ScrapeInfo struct {
	Complete   int32
	Incomplete int32
	Downloaded int32
}

type ScrapeResp struct {
	Action        int32
	TransactionID int32
	Info          []ScrapeInfo
}

func (sr *ScrapeResp) Marshall() ([]byte, error) {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, sr.Action); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, sr.TransactionID); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, sr.Info); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}
