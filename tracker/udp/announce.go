package udp

import (
	"math/rand"
	"net"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp/protocol"
)

func (u *UDPTracker) announce(announce *protocol.Announce, remote *net.UDPAddr, addr [4]byte) {
	storage.Expvar.Announces.Add(1)

	if announce.Port == 0 {
		msg := u.newClientError("bad port", announce.TransactionID, cerrFields{"addr": addr, "port": announce.Port})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	if announce.NumWant < 1 {
		announce.NumWant = config.Conf.Tracker.Numwant.Default
	} else if announce.NumWant > config.Conf.Tracker.Numwant.Limit {
		announce.NumWant = config.Conf.Tracker.Numwant.Limit
	}

	if announce.Event == protocol.Stopped {
		u.peerdb.Drop(announce.InfoHash, announce.PeerID)

		resp := protocol.AnnounceResp{
			Action:        1,
			TransactionID: announce.TransactionID,
			Interval:      int32(config.Conf.Tracker.Announce.Seconds()) + rand.Int31n(int32(config.Conf.Tracker.AnnounceFuzz.Seconds())),
			Leechers:      0,
			Seeders:       0,
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
	if announce.Event == protocol.Completed || announce.Left == 0 {
		peerComplete = true
	}

	u.peerdb.Save(addr, announce.Port, peerComplete, announce.InfoHash, announce.PeerID)
	complete, incomplete := u.peerdb.HashStats(announce.InfoHash)

	peerlist := u.peerdb.PeerListBytes(announce.InfoHash, int(announce.NumWant))
	interval := int32(config.Conf.Tracker.Announce.Seconds())
	if int32(config.Conf.Tracker.AnnounceFuzz.Seconds()) > 0 {
		interval += rand.Int31n(int32(config.Conf.Tracker.AnnounceFuzz.Seconds()))
	}

	resp := protocol.AnnounceResp{
		Action:        1,
		TransactionID: announce.TransactionID,
		Interval:      interval,
		Leechers:      int32(incomplete),
		Seeders:       int32(complete),
		Peers:         peerlist.Data,
	}
	respBytes, err := resp.Marshall()
	peerlist.Put()

	if err != nil {
		msg := u.newServerError("AnnounceResp.Marshall()", err, announce.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	storage.Expvar.AnnouncesOK.Add(1)
	u.sock.WriteToUDP(respBytes, remote)
}
