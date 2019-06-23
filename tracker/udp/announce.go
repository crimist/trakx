package udp

import (
	"bytes"
	"encoding/binary"
	"net"
	"strings"
	"time"

	"github.com/Syc0x00/Trakx/tracker"
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
	InfoHash      tracker.Hash
	PeerID        tracker.PeerID
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
	Peers         []tracker.UDPPeer
}

func (ar *AnnounceResp) Marshall() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, ar) // does this fucking work?
	return buff.Bytes()
}

func (u *UDPTracker) Announce(announce *Announce, remote *net.UDPAddr) {
	if len(announce.InfoHash) != 20 {
		e := Error{
			Action:        3,
			TransactionID: announce.TransactionID,
			ErrorString:   []byte("invalid infohash"),
		}
		u.conn.WriteToUDP(e.Marshall(), remote)
	}
	if announce.Port == 0 {
		e := Error{
			Action:        3,
			TransactionID: announce.TransactionID,
			ErrorString:   []byte("invalid port"),
		}
		u.conn.WriteToUDP(e.Marshall(), remote)
	}

	peer := tracker.Peer{
		IP:       remote.IP.String(),
		Port:     announce.Port,
		LastSeen: time.Now().Unix(),
	}
	binary.BigEndian.PutUint32(peer.Key, announce.Key) // right endian?

	if announce.Event == completed || announce.Left == 0 {
		peer.Complete = true
	}

	if strings.Contains(peer.IP, ":") {
		// return err
		return
	}

	if announce.Event == stopped {
		peer.Delete(announce.InfoHash, announce.PeerID)
		u.conn.WriteToUDP([]byte(tracker.Bye), remote)
	}

	if err := peer.Save(announce.InfoHash, announce.PeerID); err != nil {
		e := Error{
			Action:        3,
			TransactionID: announce.TransactionID,
			ErrorString:   []byte("internal server error"),
		}
		u.conn.WriteToUDP(e.Marshall(), remote)
	}

	numwant := tracker.DefaultNumwant
	if announce.NumWant != 0 {
		numwant = int(announce.NumWant)
	}
	complete, incomplete := announce.InfoHash.Complete()

	resp := AnnounceResp{
		Action:        1,
		TransactionID: announce.TransactionID,
		Interval:      tracker.AnnounceInterval,
		Leechers:      incomplete,
		Seeders:       complete,
		Peers:         announce.InfoHash.PeerListUDP(int64(numwant)),
	}
	u.conn.WriteToUDP(resp.Marshall(), remote)

	return
}
