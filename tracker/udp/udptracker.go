package udp

import (
	"encoding/binary"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type udpTracker struct {
	conn    *net.UDPConn
	avgResp time.Time
}

// GetConnCount get the number of connections in the database
func GetConnCount() int {
	return connDB.conns()
}

// WriteConns writes the connection database to file
func WriteConns() {
	connDB.write()
}

// Run runs the UDP tracker
func Run() {
	u := udpTracker{}
	connDB.load()
	rand.Seed(time.Now().UnixNano() * time.Now().Unix())

	go shared.RunOn(time.Duration(shared.Config.Database.Conn.Trim)*time.Second, connDB.trim)
	u.listen()
}

func (u *udpTracker) listen() {
	var err error

	u.conn, err = net.ListenUDP("udp4", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: shared.Config.Tracker.UDP.Port, Zone: ""})
	if err != nil {
		panic(err)
	}
	defer u.conn.Close()

	var pool sync.Pool
	pool.New = func() interface{} {
		return make([]byte, 1496, 1496)
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
			if len > 15 { // 16 = minimum connect
				u.process(b[:len], remote)
			}
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
	var addr [4]byte
	ip := remote.IP.To4()
	copy(addr[:], ip)

	connid := int64(binary.BigEndian.Uint64(data[0:8]))
	action := data[11]
	txid := int32(binary.BigEndian.Uint32(data[12:16]))

	if ip == nil {
		u.conn.WriteToUDP(newClientError("IPv6?", txid, zap.String("ip", remote.IP.String())), remote)
		return
	}

	if data[11] > 2 {
		u.conn.WriteToUDP(newClientError("bad action", txid, zap.Uint8("action", data[11]), zap.Reflect("addr", addr)), remote)
		return
	}

	if action == 0 {
		c := connect{}
		if err := c.unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("base.unmarshall()", err, txid), remote)
		}
		u.connect(&c, remote, addr)
		return
	}

	if dbID, ok := connDB.check(connid, addr); !ok && shared.Config.Tracker.UDP.CheckConnID {
		u.conn.WriteToUDP(newClientError("bad connid", txid), remote)
		if !shared.Config.Trakx.Prod {
			shared.Logger.Info("Bad connid", zap.Int64("dbID", dbID), zap.Int64("clientID", connid), zap.Reflect("ip", ip))
		}
		return
	}

	switch action {
	case 1:
		announce := announce{}
		if err := announce.unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("announce.unmarshall()", err, txid), remote)
			return
		}
		u.announce(&announce, remote, addr)

	case 2:
		scrape := scrape{}
		if err := scrape.unmarshall(data); err != nil {
			u.conn.WriteToUDP(newServerError("scrape.unmarshall()", err, txid), remote)
			return
		}
		u.scrape(&scrape, remote)
	}
}
