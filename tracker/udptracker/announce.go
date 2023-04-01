package udptracker

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"go.uber.org/zap"
)

func (tracker *Tracker) announce(udpAddr *net.UDPAddr, addrPort netip.AddrPort, transactionID int32, data []byte) {
	stats.Announces.Add(1)

	if len(data) < minimumAnnounceSize {
		tracker.sendError(udpAddr, "announce too short", transactionID)
		zap.L().Debug("client sent announce below minimum size", zap.Binary("packet", data), zap.Int("size", len(data)), zap.Any("remote", addrPort))
		return
	}

	announceRequest, err := udpprotocol.NewAnnounceRequest(data)
	if err != nil {
		tracker.sendError(udpAddr, "failed to parse announce", transactionID)
		zap.L().Debug("failed to parse clients announce packet", zap.Binary("packet", data), zap.Error(err), zap.Any("remote", addrPort))
		return
	}

	if announceRequest.Port == 0 {
		tracker.sendError(udpAddr, "invalid announce port", announceRequest.TransactionID)
		zap.L().Debug("client sent announce with invalid port", zap.Any("announce", announceRequest), zap.Uint16("port", announceRequest.Port), zap.Any("remote", udpAddr))
		return
	}

	if announceRequest.NumWant < 1 {
		announceRequest.NumWant = int32(tracker.config.DefaultNumwant)
	} else if announceRequest.NumWant > int32(tracker.config.MaximumNumwant) {
		announceRequest.NumWant = int32(tracker.config.MaximumNumwant)
	}

	interval := tracker.config.Interval
	if tracker.config.IntervalVariance > 0 {
		interval += rand.Int31n(tracker.config.IntervalVariance)
	}

	seeds, leeches := tracker.peerDB.HashStats(announceRequest.InfoHash)

	if announceRequest.Event == udpprotocol.EventStopped {
		tracker.peerDB.Drop(announceRequest.InfoHash, announceRequest.PeerID)

		marshalledResp := udpprotocol.AnnounceResponse{
			Action:        udpprotocol.ActionAnnounce,
			TransactionID: announceRequest.TransactionID,
			Interval:      interval,
			Leechers:      int32(leeches),
			Seeders:       int32(seeds),
			Peers:         []byte{},
		}
		respBytes, err := marshalledResp.Marshall()
		if err != nil {
			tracker.sendError(udpAddr, "failed to marshall announce response", announceRequest.TransactionID)
			zap.L().Error("failed to marshall announce response", zap.Error(err), zap.Any("announce", announceRequest), zap.Any("remote", udpAddr))
			return
		}

		tracker.socket.WriteToUDP(respBytes, udpAddr)
		return
	}

	peerComplete := false
	if announceRequest.Event == udpprotocol.EventCompleted || announceRequest.Left == 0 {
		peerComplete = true
	}

	tracker.peerDB.Save(addrPort.Addr(), announceRequest.Port, peerComplete, announceRequest.InfoHash, announceRequest.PeerID)

	peers4, peers6 := tracker.peerDB.PeerListBytes(announceRequest.InfoHash, uint(announceRequest.NumWant))

	marshalledResp := udpprotocol.AnnounceResponse{
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: announceRequest.TransactionID,
		Interval:      interval,
		Leechers:      int32(leeches),
		Seeders:       int32(seeds),
	}

	if addrPort.Addr().Is4() {
		marshalledResp.Peers = peers4
	} else {
		marshalledResp.Peers = peers6
	}

	respBytes, err := marshalledResp.Marshall()
	pools.Peerlists4.Put(peers4)
	pools.Peerlists6.Put(peers6)

	if err != nil {
		tracker.sendError(udpAddr, "failed to marshall announce response", announceRequest.TransactionID)
		zap.L().Error("failed to marshall announce response", zap.Error(err), zap.Any("announce", announceRequest), zap.Any("remote", udpAddr))
		return
	}

	tracker.socket.WriteToUDP(respBytes, udpAddr)
}
