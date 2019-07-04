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
	return binary.Read(bytes.NewReader(data), binary.BigEndian, c)
}

type connectResp struct {
	Action        int32
	TransactionID int32
	ConnectionID  int64
}

func (cr *connectResp) marshall() ([]byte, error) {
	var buff bytes.Buffer
	err := binary.Write(&buff, binary.BigEndian, cr)
	return buff.Bytes(), err
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
