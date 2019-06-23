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

	u.conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: shared.UDPPort, Zone: ""})
	if err != nil {
		panic(err)
	}
	defer u.conn.Close()

	buf := make([]byte, 1500)
	for {
		len, remote, err := u.conn.ReadFromUDP(buf)
		if err != nil {
			shared.Logger.Error("ReadFromUDP()", zap.Error(err))
			continue
		}
		u.Process(len, remote, buf)
	}
}

func (u *UDPTracker) Process(len int, remote *net.UDPAddr, data []byte) {
	// connectReader := bytes.NewReader(data[0:7])
	connect := Connect{}
	// binary.Read(connectReader, binary.BigEndian, &connect)
	connect.Unmarshall(data)

	// Connecting
	if connect.ConnectionID == 0x41727101980 && connect.Action == 0 {
		u.Connect(&connect, remote)
		return
	}

	var addr [4]byte
	copy(addr[:], remote.IP)

	if ok := connDB.Check(connect.ConnectionID, addr); !ok {
		e := Error{
			Action:        3,
			TransactionID: connect.TransactionID,
			ErrorString:   []byte("Invalid ConnectionID"),
		}
		e.Marshall()
	}

	switch connect.Action {
	case 1:
		announce := Announce{}
		if err := announce.Unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("announce.Unmarshall()", err, connect.TransactionID), remote)
			return
		}
		u.Announce(&announce, remote)

	case 2:
		scrape := Scrape{}
		if err := scrape.Unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("scrape.Unmarshall()", err, connect.TransactionID), remote)
			return
		}
		u.Scrape(&scrape, remote)
	}
}
