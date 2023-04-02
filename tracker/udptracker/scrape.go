package udptracker

import (
	"net"
	"net/netip"

	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"github.com/crimist/trakx/utils"
	"go.uber.org/zap"
)

const maximumScrapeHashes = 74

func (tracker *Tracker) scrape(udpAddr *net.UDPAddr, addrPort netip.AddrPort, transactionID int32, data []byte) {
	tracker.stats.Scrapes.Add(1)

	scrape, err := udpprotocol.NewScrapeRequest(data)
	if err != nil {
		tracker.sendError(udpAddr, "failed to parse scrape", transactionID)
		zap.L().Info("failed to parse clients scrape packet", zap.Binary("packet", data), zap.Error(err), zap.Any("remote", addrPort))
		return
	}

	if len(scrape.InfoHashes) > maximumScrapeHashes {
		tracker.sendError(udpAddr, "packet contains more than 74 hashes", scrape.TransactionID)
		zap.L().Debug("client sent scrape with more than 74 hashes", zap.Int("hashes", len(scrape.InfoHashes)), zap.Any("scrape", scrape), zap.Any("remote", udpAddr))
		return
	}

	resp := udpprotocol.ScrapeResponse{
		Action:        udpprotocol.ActionScrape,
		TransactionID: scrape.TransactionID,
	}

	for _, hash := range scrape.InfoHashes {
		if len(hash) != 20 {
			tracker.sendError(udpAddr, "hash "+utils.ByteToStringUnsafe(hash[0:7])+" is missized", scrape.TransactionID)
			zap.L().Debug("client sent scrape with missized hash", zap.Any("hash", hash), zap.Any("scrape", scrape), zap.Any("remote", udpAddr))
			return
		}

		seeds, leeches := tracker.peerDB.TorrentStats(hash)
		info := udpprotocol.ScrapeResponseInfo{
			Complete:   int32(seeds),
			Incomplete: int32(leeches),
			Downloaded: -1,
		}
		resp.Info = append(resp.Info, info)
	}

	marshalledResp, err := resp.Marshall()
	if err != nil {
		tracker.sendError(udpAddr, "failed to marshall scrape response", scrape.TransactionID)
		zap.L().Error("failed to marshall scrape response", zap.Error(err), zap.Any("scrape", scrape), zap.Any("remote", udpAddr))
		return
	}

	tracker.socket.WriteToUDP(marshalledResp, udpAddr)
}
