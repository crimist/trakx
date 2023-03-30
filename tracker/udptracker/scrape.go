package udptracker

import (
	"net"

	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/udptracker/protocol"
)

func (u *Tracker) scrape(scrape *protocol.Scrape, remote *net.UDPAddr) {
	stats.Scrapes.Add(1)

	if len(scrape.InfoHashes) > 74 {
		msg := u.newClientError("74 hashes max", scrape.TransactionID)
		u.socket.WriteToUDP(msg, remote)
		return
	}

	resp := protocol.ScrapeResp{
		Action:        protocol.ActionScrape,
		TransactionID: scrape.TransactionID,
	}

	for _, hash := range scrape.InfoHashes {
		if len(hash) != 20 {
			msg := u.newClientError("bad hash", scrape.TransactionID)
			u.socket.WriteToUDP(msg, remote)
			return
		}

		complete, incomplete := u.peerDB.HashStats(hash)
		info := protocol.ScrapeInfo{
			Complete:   int32(complete),
			Incomplete: int32(incomplete),
			Downloaded: -1,
		}
		resp.Info = append(resp.Info, info)
	}

	respBytes, err := resp.Marshall()
	if err != nil {
		msg := u.newServerError("ScrapeResp.Marshall()", err, scrape.TransactionID)
		u.socket.WriteToUDP(msg, remote)
		return
	}

	u.socket.WriteToUDP(respBytes, remote)
}
