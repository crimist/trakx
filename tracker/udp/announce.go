package udp

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync/atomic"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
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
	InfoHash      shared.Hash
	PeerID        shared.PeerID
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

func (u *udpTracker) announce(announce *announce, remote *net.UDPAddr, addr [4]byte) {
	atomic.AddInt64(&shared.ExpvarAnnounces, 1)

	if announce.Port == 0 {
		u.conn.WriteToUDP(newClientError("bad port", announce.TransactionID, zap.Reflect("addr", addr), zap.Uint16("port", announce.Port)), remote)
		return
	}

	peer := shared.Peer{
		IP:       addr,
		Port:     announce.Port,
		LastSeen: time.Now().Unix(),
	}

	if announce.Event == completed || announce.Left == 0 {
		peer.Complete = true
	}
	if announce.NumWant < 1 || announce.NumWant > shared.Config.Tracker.Numwant.Max {
		announce.NumWant = shared.Config.Tracker.Numwant.Default
	}

	if announce.Event == stopped {
		shared.PeerDB.Drop(&peer, &announce.InfoHash, &announce.PeerID)

		atomic.AddInt64(&shared.ExpvarAnnouncesOK, 1)
		if shared.Config.Tracker.StoppedMsg != "" {
			u.conn.WriteToUDP([]byte(shared.Config.Tracker.StoppedMsg), remote)
		}
		return
	}

	shared.PeerDB.Save(&peer, &announce.InfoHash, &announce.PeerID)

	complete, incomplete := shared.PeerDB.HashStats(&announce.InfoHash)

	resp := announceResp{
		Action:        1,
		TransactionID: announce.TransactionID,
		Interval:      shared.Config.Tracker.AnnounceInterval,
		Leechers:      incomplete,
		Seeders:       complete,
		Peers:         shared.PeerDB.PeerListBytes(&announce.InfoHash, int(announce.NumWant)),
	}
	respBytes, err := resp.marshall()
	if err != nil {
		u.conn.WriteToUDP(newServerError("AnnounceResp.Marshall()", err, announce.TransactionID), remote)
		return
	}

	atomic.AddInt64(&shared.ExpvarAnnouncesOK, 1)
	u.conn.WriteToUDP(respBytes, remote)
	return
}
