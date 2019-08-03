package udp

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync/atomic"

	"github.com/syc0x00/trakx/tracker/shared"
)

type scrape struct {
	Base     connect
	InfoHash []shared.Hash
}

func (s *scrape) unmarshall(data []byte) error {
	if err := binary.Read(bytes.NewReader(data[:16]), binary.BigEndian, &s.Base); err != nil {
		return err
	}

	s.InfoHash = make([]shared.Hash, (len(data)-16)/20)
	return binary.Read(bytes.NewReader(data[16:]), binary.BigEndian, &s.InfoHash)
}

type scrapeInfo struct {
	Complete   int32
	Downloaded int32
	Incomplete int32
}

type scrapeResp struct {
	Action        int32
	TransactionID int32
	Info          []scrapeInfo
}

func (sr *scrapeResp) marshall() ([]byte, error) {
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

func (u *UDPTracker) scrape(scrape *scrape, remote *net.UDPAddr) {
	atomic.AddInt64(&shared.Expvar.Scrapes, 1)

	if len(scrape.InfoHash) > 74 {
		msg := u.newClientError("74 hashes max", scrape.Base.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	resp := scrapeResp{
		Action:        2,
		TransactionID: scrape.Base.TransactionID,
	}

	for _, hash := range scrape.InfoHash {
		if len(hash) != 20 {
			msg := u.newClientError("bad hash", scrape.Base.TransactionID)
			u.sock.WriteToUDP(msg, remote)
			return
		}

		complete, incomplete := shared.PeerDB.HashStats(&hash)
		info := scrapeInfo{
			Complete:   complete,
			Downloaded: -1,
			Incomplete: incomplete,
		}
		resp.Info = append(resp.Info, info)
	}

	respBytes, err := resp.marshall()
	if err != nil {
		msg := u.newServerError("ScrapeResp.Marshall()", err, scrape.Base.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	u.sock.WriteToUDP(respBytes, remote)
	atomic.AddInt64(&shared.Expvar.ScrapesOK, 1)
	return
}
