package udp

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"

	"github.com/syc0x00/trakx/tracker/database"
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

func (u *UDPTracker) connect(connect *connect, remote *net.UDPAddr, addr [4]byte) {
	database.AddExpval(&database.Expvar.Connects, 1)

	id := rand.Int63()
	u.conndb.add(id, addr)

	resp := connectResp{
		Action:        0,
		TransactionID: connect.TransactionID,
		ConnectionID:  id,
	}

	respBytes, err := resp.marshall()
	if err != nil {
		msg := u.newServerError("ConnectResp.Marshall()", err, connect.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	database.AddExpval(&database.Expvar.ConnectsOK, 1)
	u.sock.WriteToUDP(respBytes, remote)
	return
}
