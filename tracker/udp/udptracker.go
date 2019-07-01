package udp

import (
	"math/rand"
	"net"
	"time"

	"github.com/Syc0x00/Trakx/tracker/shared"
	"go.uber.org/zap"
)

// https://www.libtorrent.org/udp_tracker_protocol.html

type UDPTracker struct {
	conn    *net.UDPConn
	avgResp time.Time
}

func (u *UDPTracker) Trimmer() {
	for c := time.Tick(1 * time.Minute); ; <-c {
		connDB.Trim()
	}
}

func (u *UDPTracker) Listen() {
	var err error
	rand.Seed(time.Now().UnixNano() * time.Now().Unix())
	connDB = make(UDPConnDB)

	u.conn, err = net.ListenUDP("udp4", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: shared.UDPPort, Zone: ""})
	if err != nil {
		panic(err)
	}
	defer u.conn.Close()

	buf := make([]byte, 1496)
	for {
		len, remote, err := u.conn.ReadFromUDP(buf)
		if err != nil {
			shared.Logger.Error("ReadFromUDP()", zap.Error(err))
			continue
		}
		go u.Process(len, remote, buf)
	}
}

func (u *UDPTracker) Process(len int, remote *net.UDPAddr, data []byte) {
	connect := Connect{}
	connect.Unmarshall(data)
	var addr [4]byte
	ip := remote.IP.To4()
	copy(addr[:], ip)

	if ip == nil {
		u.conn.WriteToUDP(newClientError("how did you use ipv6???", connect.TransactionID), remote)
		return
	}

	// Connecting
	if connect.ConnectionID == 0x41727101980 && connect.Action == 0 {
		u.Connect(&connect, remote, addr)
		return
	}

	if ok := connDB.Check(connect.ConnectionID, addr); ok == false {
		shared.ExpvarClienterrs++
		e := Error{
			Action:        3,
			TransactionID: connect.TransactionID,
			ErrorString:   []byte("bad connid"),
		}
		respBytes, err := e.Marshall()
		if err != nil {
			u.conn.WriteToUDP(newServerError("Error.Marshall()", err, connect.TransactionID), remote)
			return
		}
		u.conn.WriteToUDP(respBytes, remote)
	}

	switch connect.Action {
	case 1:
		announce := Announce{}
		if err := announce.Unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("announce.Unmarshall()", err, connect.TransactionID), remote)
			return
		}
		u.Announce(&announce, remote, addr)

	case 2:
		scrape := Scrape{}
		if err := scrape.Unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("scrape.Unmarshall()", err, connect.TransactionID), remote)
			return
		}
		u.Scrape(&scrape, remote)
	default:
		u.conn.WriteToUDP(newClientError("bad action", connect.TransactionID), remote)
	}
}
