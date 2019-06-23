package udp

import (
	"bytes"
	"encoding/binary"
	"net"

	"github.com/Syc0x00/Trakx/tracker/shared"
)

type Scrape struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
	InfoHash      []shared.Hash
}

func (s *Scrape) Unmarshall(data []byte) error {
	reader := bytes.NewReader(data)
	return binary.Read(reader, binary.BigEndian, s)
}

type scrapeInfo struct {
	Complete   int32
	Downloaded int32
	Incomplete int32
}

type ScrapeResp struct {
	Action        int32
	TransactionID int32
	Info          []scrapeInfo
}

func (sr *ScrapeResp) Marshall() ([]byte, error) {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, sr); err != nil { // does this fucking work?
		return nil, err
	}
	return buff.Bytes(), nil
}

func (u *UDPTracker) Scrape(scrape *Scrape, remote *net.UDPAddr) {
	shared.ExpvarScrapes++

	resp := ScrapeResp{
		Action:        2,
		TransactionID: scrape.TransactionID,
	}

	for _, hash := range scrape.InfoHash {
		if len(hash) != 20 {
			u.conn.WriteToUDP(newClientError("invalid infohash", scrape.TransactionID), remote)
			return
		}

		complete, incomplete := hash.Complete()
		info := scrapeInfo{
			Complete:   complete,
			Downloaded: -1,
			Incomplete: incomplete,
		}
		resp.Info = append(resp.Info, info)
	}

	respBytes, err := resp.Marshall()
	if err != nil {
		u.conn.WriteToUDP(newServerError("ScrapeResp.Marshall()", err, scrape.TransactionID), remote)
		return
	}

	u.conn.WriteToUDP(respBytes, remote)
	return
}
