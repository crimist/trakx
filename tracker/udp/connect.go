package udp

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
)

type Connect struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
}

func (c *Connect) Unmarshall(data []byte) error {
	reader := bytes.NewReader(data)
	return binary.Read(reader, binary.BigEndian, c)
}

type ConnectResp struct {
	Action        int32
	TransactionID int32
	ConnectionID  int64
}

func (cr *ConnectResp) Marshall() []byte {
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.BigEndian, cr.Action)
	binary.Write(buff, binary.BigEndian, cr.TransactionID)
	binary.Write(buff, binary.BigEndian, cr.ConnectionID)
	return buff.Bytes()
}

func (u *UDPTracker) Connect(connect *Connect, remote *net.UDPAddr) {
	id := rand.Int63()
	var addr [4]byte
	copy(addr[:], remote.IP)

	connDB.Add(id, addr)

	resp := ConnectResp{
		Action:        0,
		TransactionID: connect.TransactionID,
		ConnectionID:  id,
	}

	u.conn.WriteToUDP(resp.Marshall(), remote)

	return
}
