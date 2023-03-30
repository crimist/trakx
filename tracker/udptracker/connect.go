package udptracker

import (
	"math/rand"
	"net"
	"net/netip"

	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/udptracker/protocol"
)

func (u *Tracker) connect(connect protocol.Connect, remote *net.UDPAddr, addr netip.AddrPort) {
	stats.Connects.Add(1)

	id := rand.Int63()
	u.connDB.add(id, addr)

	resp := protocol.ConnectResp{
		Action:        protocol.ActionConnect,
		TransactionID: connect.TransactionID,
		ConnectionID:  id,
	}

	respBytes, err := resp.Marshall()
	if err != nil {
		msg := u.newServerError("ConnectResp.Marshall()", err, connect.TransactionID)
		u.socket.WriteToUDP(msg, remote)
		return
	}

	u.socket.WriteToUDP(respBytes, remote)
}
