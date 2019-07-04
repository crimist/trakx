package udp

import (
	"math/rand"
	"net"
	"time"
	"sync"

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

	var pool sync.Pool
	pool.New = func() interface{} {
		return make([]byte, 2048, 2048)
	}

	for {
		b := pool.Get().([]byte)
		len, remote, err := u.conn.ReadFromUDP(b)
		if err != nil {
			shared.Logger.Error("ReadFromUDP()", zap.Error(err))
			pool.Put(b)
			continue
		}
		go func() {
			u.process(b[:len], remote)

			// optimized zero
			b = b[:cap(b)]
			for i := range b {
				b[i] = 0
			}
			pool.Put(b)
		}()
	}
}

func (u *udpTracker) process(data []byte, remote *net.UDPAddr) {
	base := connect{}
	var addr [4]byte
	ip := remote.IP.To4()
	copy(addr[:], ip)

	if ip == nil {
		u.conn.WriteToUDP(newClientError("how did you use ipv6???", base.TransactionID, zap.ByteString("ip", remote.IP)), remote)
		return
	}

	err := base.unmarshall(data)
	if err != nil {
		u.conn.WriteToUDP(newServerError("base.unmarshall()", err, base.TransactionID), remote)
	}

	if base.Action == 0 {
		u.connect(&base, remote, addr)
		return
	}

	if dbID, ok := connDB.check(base.ConnectionID, addr); !ok {
		u.conn.WriteToUDP(newClientError("Connection ID missmatch.", base.TransactionID, zap.Int64("dbID", dbID)), remote)
		return
	}

	switch base.Action {
	case 1:
		announce := announce{}
		if err := announce.unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("announce.unmarshall()", err, base.TransactionID), remote)
			return
		}
		u.announce(&announce, remote, addr)

	case 2:
		scrape := scrape{}
		if err := scrape.unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("scrape.unmarshall()", err, base.TransactionID), remote)
			return
		}
		u.scrape(&scrape, remote)
	default:
		u.conn.WriteToUDP(newClientError("bad action", base.TransactionID, zap.Int32("action", base.Action)), remote)
	}
}
