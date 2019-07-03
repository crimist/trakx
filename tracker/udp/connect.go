package udp

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"

	"github.com/Syc0x00/Trakx/tracker/shared"
)

type connect struct {
	ConnectionID  int64
	Action        int32
	TransactionID int32
}

func (c *connect) unmarshall(data []byte) error {
	reader := bytes.NewReader(data)
	return binary.Read(reader, binary.BigEndian, c)
}

type connectResp struct {
	Action        int32
	TransactionID int32
	ConnectionID  int64
}

func (cr *connectResp) marshall() ([]byte, error) {
	buff := new(bytes.Buffer)
	if err := binary.Write(buff, binary.BigEndian, cr.Action); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, cr.TransactionID); err != nil {
		return nil, err
	}
	if err := binary.Write(buff, binary.BigEndian, cr.ConnectionID); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (u *udpTracker) connect(connect *connect, remote *net.UDPAddr, addr [4]byte) {
	shared.ExpvarConnects++
	id := rand.Int63()
	connDB.add(id, addr)

	resp := connectResp{
		Action:        0,
		TransactionID: connect.TransactionID,
		ConnectionID:  id,
	}

	respBytes, err := resp.marshall()
	if err != nil {
		u.conn.WriteToUDP(newServerError("ConnectResp.Marshall()", err, connect.TransactionID), remote)
		return
	}

	shared.ExpvarConnectsOK++
	u.conn.WriteToUDP(respBytes, remote)
	return
}
