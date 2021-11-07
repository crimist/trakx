package udp

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
)

type event int32

const (
	none      event = 0
	completed event = 1
	started   event = 2
	stopped   event = 3
)

type announce struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
	InfoHash      storage.Hash
	PeerID        storage.PeerID
	Downloaded    int64
	Left          int64
	Uploaded      int64
	Event         event
	IP            uint32
	Key           uint32
	NumWant       int32
	Port          uint16
	// Extensions    uint16
}

func (a *announce) unmarshall(data []byte) error {
	return binary.Read(bytes.NewReader(data), binary.BigEndian, a)
}

type announceResp struct {
	Action        int32
	TransactionID int32
	Interval      int32
	Leechers      int32
	Seeders       int32
	Peers         []byte
}

func (ar *announceResp) marshall() ([]byte, error) {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, ar.Action); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, ar.TransactionID); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, ar.Interval); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, ar.Leechers); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, ar.Seeders); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, ar.Peers); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (u *UDPTracker) announce(announce *announce, remote *net.UDPAddr, addr [4]byte) {
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

	if announce.Event == stopped {
		u.peerdb.Drop(announce.InfoHash, announce.PeerID)

		resp := announceResp{
			Action:        1,
			TransactionID: announce.TransactionID,
			Interval:      config.Conf.Tracker.Announce + rand.Int31n(config.Conf.Tracker.AnnounceFuzz),
			Leechers:      0,
			Seeders:       0,
			Peers:         []byte{},
		}
		respBytes, err := resp.marshall()
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
	if announce.Event == completed || announce.Left == 0 {
		peerComplete = true
	}

	u.peerdb.Save(addr, announce.Port, peerComplete, announce.InfoHash, announce.PeerID)
	complete, incomplete := u.peerdb.HashStats(announce.InfoHash)

	peerlist := u.peerdb.PeerListBytes(announce.InfoHash, int(announce.NumWant))
	resp := announceResp{
		Action:        1,
		TransactionID: announce.TransactionID,
		Interval:      config.Conf.Tracker.Announce + rand.Int31n(config.Conf.Tracker.AnnounceFuzz),
		Leechers:      int32(incomplete),
		Seeders:       int32(complete),
		Peers:         peerlist.Data,
	}
	respBytes, err := resp.marshall()
	peerlist.Put()
	if err != nil {
		msg := u.newServerError("AnnounceResp.Marshall()", err, announce.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	storage.Expvar.AnnouncesOK.Add(1)
	u.sock.WriteToUDP(respBytes, remote)
	return
}
