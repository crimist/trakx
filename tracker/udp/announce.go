package udp

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

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
	storage.AddExpval(&storage.Expvar.Announces, 1)

	if announce.Port == 0 {
		msg := u.newClientError("bad port", announce.TransactionID, cerrFields{"addr": addr, "port": announce.Port})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	peer := storage.Peer{
		IP:       addr,
		Port:     announce.Port,
		LastSeen: time.Now().Unix(),
	}

	if announce.Event == completed || announce.Left == 0 {
		peer.Complete = true
	}
	if announce.NumWant < 1 || announce.NumWant > u.conf.Tracker.Numwant.Max {
		announce.NumWant = u.conf.Tracker.Numwant.Default
	}

	if announce.Event == stopped {
		u.peerdb.Drop(&peer, &announce.InfoHash, &announce.PeerID)

		storage.AddExpval(&storage.Expvar.AnnouncesOK, 1)
		if u.conf.Tracker.StoppedMsg != "" {
			u.sock.WriteToUDP([]byte(u.conf.Tracker.StoppedMsg), remote)
		}
		return
	}

	u.peerdb.Save(&peer, &announce.InfoHash, &announce.PeerID)

	complete, incomplete := u.peerdb.HashStats(&announce.InfoHash)

	resp := announceResp{
		Action:        1,
		TransactionID: announce.TransactionID,
		Interval:      u.conf.Tracker.Announce,
		Leechers:      incomplete,
		Seeders:       complete,
		Peers:         u.peerdb.PeerListBytes(&announce.InfoHash, int(announce.NumWant)),
	}
	respBytes, err := resp.marshall()
	if err != nil {
		msg := u.newServerError("AnnounceResp.Marshall()", err, announce.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	storage.AddExpval(&storage.Expvar.AnnouncesOK, 1)
	u.sock.WriteToUDP(respBytes, remote)
	return
}
