package udp

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/crimist/trakx/internal/storage"
	"github.com/crimist/trakx/internal/tracker/udp/udpprotocol"
	"go.uber.org/zap"
)

var (
	fatalInvalidPort = []byte("invalid announce port")
)

func (tracker *Tracker) announce(udpAddr *net.UDPAddr, addrPort netip.AddrPort, transactionID int32, data []byte, socket *net.UDPConn) {
	tracker.collector.Announce()

	if len(data) < minimumAnnounceSize {
		tracker.error(udpAddr, []byte("announce too short"), transactionID, socket)
		zap.L().Debug("client sent announce below minimum size", zap.Binary("packet", data), zap.Int("size", len(data)), zap.Any("remote", addrPort))
		return
	}

	announceRequest, err := udpprotocol.NewAnnounceRequest(data)
	if err != nil {
		tracker.error(udpAddr, []byte("failed to parse announce"), transactionID, socket)
		zap.L().Debug("failed to parse clients announce packet", zap.Binary("packet", data), zap.Error(err), zap.Any("remote", addrPort))
		return
	}

	if announceRequest.Port == 0 {
		tracker.error(udpAddr, fatalInvalidPort, announceRequest.TransactionID, socket)
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
		interval += uint(rand.Int63n(int64(tracker.config.IntervalVariance)))
	}

	if announceRequest.Event == udpprotocol.EventStopped {
		tracker.peerDB.PeerRemove(announceRequest.InfoHash, announceRequest.PeerID)
		seeds, leeches := tracker.peerDB.TorrentStats(announceRequest.InfoHash)

		marshalledResp := udpprotocol.AnnounceResponse{
			Action:        udpprotocol.ActionAnnounce,
			TransactionID: announceRequest.TransactionID,
			Interval:      int32(interval),
			Leeches:       int32(leeches),
			Seeds:         int32(seeds),
			Peers:         []byte{},
		}
		respBytes, err := marshalledResp.Marshal()
		if err != nil {
			tracker.error(udpAddr, []byte("failed to marshall announce response"), announceRequest.TransactionID, socket)
			zap.L().Error("failed to marshall announce response", zap.Error(err), zap.Any("announce", announceRequest), zap.Any("remote", udpAddr))
			return
		}

		socket.WriteToUDP(respBytes, udpAddr)
		return
	}

	peerComplete := false
	if announceRequest.Event == udpprotocol.EventCompleted || announceRequest.Left == 0 {
		peerComplete = true
	}

	tracker.peerDB.PeerAdd(announceRequest.InfoHash, announceRequest.PeerID, addrPort.Addr(), announceRequest.Port, peerComplete)
	seeds, leeches := tracker.peerDB.TorrentStats(announceRequest.InfoHash)

	var ipversion storage.IPVersion
	if addrPort.Addr().Is4() {
		ipversion = storage.IPv4
	} else {
		ipversion = storage.IPv6
	}

	peers := tracker.peerDB.TorrentPeersCompact(announceRequest.InfoHash, uint(announceRequest.NumWant), ipversion)

	response := udpprotocol.AnnounceResponse{
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: announceRequest.TransactionID,
		Interval:      int32(interval),
		Leeches:       int32(leeches),
		Seeds:         int32(seeds),
	}

	if ipversion == storage.IPv4 {
		response.Peers = peers.V4
	} else {
		response.Peers = peers.V6
	}

	respBytes, err := response.Marshal()
	peers.Release()

	if err != nil {
		tracker.error(udpAddr, []byte("failed to marshall announce response"), announceRequest.TransactionID, socket)
		zap.L().Error("failed to marshall announce response", zap.Error(err), zap.Any("announce", announceRequest), zap.Any("remote", udpAddr))
		return
	}

	socket.WriteToUDP(respBytes, udpAddr)
}
