package udp

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"github.com/Syc0x00/Trakx/tracker/shared"
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
	Extensions    uint16
}

func (a *announce) unmarshall(data []byte) error {
	reader := bytes.NewReader(data)
	return binary.Read(reader, binary.BigEndian, a)
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
	shared.ExpvarAnnounces++

	if announce.Port == 0 {
		u.conn.WriteToUDP(newClientError("bad port", announce.TransactionID), remote)
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
	if announce.NumWant < 1 || announce.NumWant > shared.MaxNumwant {
		announce.NumWant = shared.DefaultNumwant
	}

	if announce.Event == stopped {
		peer.Delete(announce.InfoHash, announce.PeerID)
		shared.ExpvarAnnouncesOK++
		u.conn.WriteToUDP([]byte(shared.Bye), remote)
		return
	}

	peer.Save(announce.InfoHash, announce.PeerID)

	complete, incomplete := announce.InfoHash.Complete()

	resp := announceResp{
		Action:        1,
		TransactionID: announce.TransactionID,
		Interval:      shared.AnnounceInterval,
		Leechers:      incomplete,
		Seeders:       complete,
		Peers:         announce.InfoHash.PeerListBytes(int(announce.NumWant)),
	}
	respBytes, err := resp.marshall()
	if err != nil {
		u.conn.WriteToUDP(newServerError("AnnounceResp.Marshall()", err, announce.TransactionID), remote)
		return
	}

	shared.ExpvarAnnouncesOK++
	u.conn.WriteToUDP(respBytes, remote)
	return
}
