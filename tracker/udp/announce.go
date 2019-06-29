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

type Announce struct {
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

func (a *Announce) Unmarshall(data []byte) error {
	reader := bytes.NewReader(data)
	return binary.Read(reader, binary.BigEndian, a)
}

type AnnounceResp struct {
	Action        int32
	TransactionID int32
	Interval      int32
	Leechers      int32
	Seeders       int32
	Peers         []byte
}

func (ar *AnnounceResp) Marshall() ([]byte, error) {
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

func (u *UDPTracker) Announce(announce *Announce, remote *net.UDPAddr) {
	shared.ExpvarAnnounces++
	ip := remote.IP.To4()

	if ip == nil {
		u.conn.WriteToUDP(newClientError("ipv6 unsupported", announce.TransactionID), remote)
		return
	}
	if len(announce.InfoHash) != 20 {
		u.conn.WriteToUDP(newClientError("invalid infohash", announce.TransactionID), remote)
		return
	}
	if announce.Port == 0 {
		u.conn.WriteToUDP(newClientError("invalid port", announce.TransactionID), remote)
		return
	}

	peer := shared.Peer{
		Port:     announce.Port,
		LastSeen: time.Now().Unix(),
	}
	copy(peer.IP[:], ip)

	if announce.Event == completed || announce.Left == 0 {
		peer.Complete = true
	}
	if announce.NumWant == -1 || announce.NumWant > shared.MaxNumwant {
		announce.NumWant = shared.DefaultNumwant
	}

	if announce.Event == stopped {
		peer.Delete(announce.InfoHash, announce.PeerID)
		u.conn.WriteToUDP([]byte(shared.Bye), remote)
		return
	}

	peer.Save(announce.InfoHash, announce.PeerID)

	complete, incomplete := announce.InfoHash.Complete()

	resp := AnnounceResp{
		Action:        1,
		TransactionID: announce.TransactionID,
		Interval:      shared.AnnounceInterval,
		Leechers:      incomplete,
		Seeders:       complete,
		Peers:         announce.InfoHash.PeerListBytes(int64(announce.NumWant)),
	}
	respBytes, err := resp.Marshall()
	if err != nil {
		u.conn.WriteToUDP(newServerError("AnnounceResp.Marshall()", err, announce.TransactionID), remote)
		return
	}

	u.conn.WriteToUDP(respBytes, remote)
	return
}
