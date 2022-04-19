package udp

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp/protocol"
)

func (u *UDPTracker) announce(announce *protocol.Announce, remote *net.UDPAddr, addrPort netip.AddrPort) {
	storage.Expvar.Announces.Add(1)

	if announce.Port == 0 {
		msg := u.newClientError("bad port", announce.TransactionID, cerrFields{"addrPort": addrPort, "port": announce.Port})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	if announce.NumWant < 1 {
		announce.NumWant = int32(config.Conf.Tracker.Numwant.Default)
	} else if announce.NumWant > int32(config.Conf.Tracker.Numwant.Limit) {
		announce.NumWant = int32(config.Conf.Tracker.Numwant.Limit)
	}

	if announce.Event == protocol.EventStopped {
		u.peerdb.Drop(announce.InfoHash, announce.PeerID)

		resp := protocol.AnnounceResp{
			Action:        protocol.ActionAnnounce,
			TransactionID: announce.TransactionID,
			Interval:      -1,
			Leechers:      -1,
			Seeders:       -1,
			Peers:         []byte{},
		}
		respBytes, err := resp.Marshall()
		if err != nil {
			msg := u.newServerError("AnnounceResp.Marshall()", err, announce.TransactionID)
			u.sock.WriteToUDP(msg, remote)
			return
		}

		storage.Expvar.AnnouncesOK.Add(1)
		u.sock.WriteToUDP(respBytes, remote)
		return
	}

	peerComplete := false
	if announce.Event == protocol.EventCompleted || announce.Left == 0 {
		peerComplete = true
	}

	u.peerdb.Save(addrPort.Addr(), announce.Port, peerComplete, announce.InfoHash, announce.PeerID)

	complete, incomplete := u.peerdb.HashStats(announce.InfoHash)
	peers4, peers6 := u.peerdb.PeerListBytes(announce.InfoHash, uint(announce.NumWant))
	interval := int32(config.Conf.Tracker.Announce.Seconds())
	if int32(config.Conf.Tracker.AnnounceFuzz.Seconds()) > 0 {
		interval += rand.Int31n(int32(config.Conf.Tracker.AnnounceFuzz.Seconds()))
	}

	resp := protocol.AnnounceResp{
		Action:        protocol.ActionAnnounce,
		TransactionID: announce.TransactionID,
		Interval:      interval,
		Leechers:      int32(incomplete),
		Seeders:       int32(complete),
	}

	// ipv4 or ipv6 response
	if remote.AddrPort().Addr().Is6() {
		resp.Peers = peers6.Data
	} else {
		resp.Peers = peers4.Data
	}

	respBytes, err := resp.Marshall()
	peers4.Put()
	peers6.Put()

	if err != nil {
		msg := u.newServerError("AnnounceResp.Marshall()", err, announce.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	storage.Expvar.AnnouncesOK.Add(1)
	u.sock.WriteToUDP(respBytes, remote)
}
