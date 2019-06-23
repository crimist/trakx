package udp

import (
	"bytes"
	"encoding/binary"
	"net"
)

type Announce struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
	InfoHash      [20]byte
	PeerID        [20]byte
	Downloaded    int64
	Left          int64
	Uploaded      int64
	Event         int32
	IP            uint32
	Key           uint32
	NumWant       int32
	Port          uint16
	Extensions    uint16
}

func (a *Announce) Marshall(data []byte) error {
	reader := bytes.NewReader(data)
	return binary.Read(reader, binary.BigEndian, a)
}

type AnnounceResp struct {
	Action        int32
	TransactionID int32
	Interval      int32
	Leechers      int32
	Seeders       int32
	Peers         []struct {
		IP   int32
		Port uint16
	}
}

func (ar *AnnounceResp) Unmarshall() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, ar) // does this fucking work?
	return buff.Bytes()
}

func (u *UDPTracker) Announce(announce *Announce, remote *net.UDPAddr) {
	return
}
