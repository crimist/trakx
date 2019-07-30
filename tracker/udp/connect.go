package udp

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"net"
	"sync/atomic"

	"github.com/syc0x00/trakx/tracker/shared"
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
	atomic.AddInt64(&shared.ExpvarConnects, 1)
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

	atomic.AddInt64(&shared.ExpvarConnectsOK, 1)
	u.conn.WriteToUDP(respBytes, remote)
	return
}
