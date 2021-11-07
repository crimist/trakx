package udp

import (
	"net"

	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp/protocol"
)

func (u *UDPTracker) scrape(scrape *protocol.Scrape, remote *net.UDPAddr) {
	storage.Expvar.Scrapes.Add(1)

	if len(scrape.InfoHash) > 74 {
		msg := u.newClientError("74 hashes max", scrape.Base.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	resp := protocol.ScrapeResp{
		Action:        2,
		TransactionID: scrape.Base.TransactionID,
	}

	for _, hash := range scrape.InfoHash {
		if len(hash) != 20 {
			msg := u.newClientError("bad hash", scrape.Base.TransactionID)
			u.sock.WriteToUDP(msg, remote)
			return
		}

		complete, incomplete := u.peerdb.HashStats(hash)
		info := protocol.ScrapeInfo{
			Complete:   int32(complete),
			Incomplete: int32(incomplete),
			Downloaded: -1,
		}
		resp.Info = append(resp.Info, info)
	}

	respBytes, err := resp.Marshall()
	if err != nil {
		msg := u.newServerError("ScrapeResp.Marshall()", err, scrape.Base.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	u.sock.WriteToUDP(respBytes, remote)
	storage.Expvar.ScrapesOK.Add(1)
}
