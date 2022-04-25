package udp

import (
	"math/rand"
	"net"

	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp/protocol"
)

func (u *UDPTracker) connect(connect *protocol.Connect, remote *net.UDPAddr, addr connAddr) {
	storage.Expvar.Connects.Add(1)

	id := rand.Int63()
	u.conndb.add(id, addr)

	resp := protocol.ConnectResp{
		Action:        protocol.ActionConnect,
		TransactionID: connect.TransactionID,
		ConnectionID:  id,
	}

	respBytes, err := resp.Marshall()
	if err != nil {
		msg := u.newServerError("ConnectResp.Marshall()", err, connect.TransactionID)
		u.sock.WriteToUDP(msg, remote)
		return
	}

	storage.Expvar.ConnectsOK.Add(1)
	u.sock.WriteToUDP(respBytes, remote)
}
