package udp

import (
	"math/rand"
	"net"
	"time"

	"github.com/Syc0x00/Trakx/tracker/shared"
	"go.uber.org/zap"
)

type udpTracker struct {
	conn    *net.UDPConn
	avgResp time.Time
}

// Run runs the UDP tracker
func Run(trimInterval time.Duration) {
	u := udpTracker{}
	connDB = make(udpConnDB)
	rand.Seed(time.Now().UnixNano() * time.Now().Unix())

	go shared.RunOn(trimInterval, connDB.trim)
	u.listen()
}

func (u *udpTracker) listen() {
	var err error

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
		go u.process(len, remote, buf)
	}
}

func (u *udpTracker) process(len int, remote *net.UDPAddr, data []byte) {
	base := connect{}
	base.unmarshall(data)
	var addr [4]byte
	ip := remote.IP.To4()
	copy(addr[:], ip)

	if ip == nil {
		u.conn.WriteToUDP(newClientError("how did you use ipv6???", base.TransactionID, zap.ByteString("ip", remote.IP)), remote)
		return
	}

	if base.Action == 0 { // connect.ConnectionID == 0x41727101980
		u.connect(&base, remote, addr)
		return
	}

	if ok := connDB.check(base.ConnectionID, addr); !ok {
		u.conn.WriteToUDP(newClientError("bad connid", base.TransactionID), remote)
		return
	}

	switch base.Action {
	case 1:
		announce := announce{}
		if err := announce.unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("announce.Unmarshall()", err, base.TransactionID), remote)
			return
		}
		u.announce(&announce, remote, addr)

	case 2:
		scrape := scrape{}
		if err := scrape.unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("scrape.Unmarshall()", err, base.TransactionID), remote)
			return
		}
		u.scrape(&scrape, remote)
	default:
		u.conn.WriteToUDP(newClientError("bad action", base.TransactionID, zap.Int32("action", base.Action)), remote)
	}
}
