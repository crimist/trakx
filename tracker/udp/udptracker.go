package udp

import (
	"encoding/binary"
	"net"
	"sync"
	"time"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

type UDPTracker struct {
	sock   *net.UDPConn
	conndb *connectionDatabase

	conf   *shared.Config
	logger *zap.Logger
	peerdb storage.Database
}

// Init creates are runs the UDP tracker
func (u *UDPTracker) Init(conf *shared.Config, logger *zap.Logger, peerdb storage.Database) {
	u.conndb = newConnectionDatabase(conf.Database.Conn.Timeout, conf.Database.Conn.Filename, logger)
	u.conf = conf
	u.logger = logger
	u.peerdb = peerdb

	go shared.RunOn(time.Duration(conf.Database.Conn.Trim)*time.Second, u.conndb.trim)
}

func (u *UDPTracker) Serve() {
	var err error

	u.sock, err = net.ListenUDP("udp4", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: u.conf.Tracker.UDP.Port, Zone: ""})
	if err != nil {
		panic(err)
	}
	defer u.sock.Close()

	pool := sync.Pool{
		New: func() interface{} { return make([]byte, 1496, 1496) }, // 1496 is max size of a scrape with 20 hashes
	}

	for i := 0; i < u.conf.Tracker.UDP.Threads; i++ {
		go func() {
			for {
				data := pool.Get().([]byte)
				l, remote, err := u.sock.ReadFromUDP(data)
				if err != nil {
					u.logger.Error("ReadFromUDP()", zap.Error(err))
					pool.Put(data)
					continue
				}

				if l > 15 { // 16 = minimum connect
					u.process(data[:l], remote)
				}

				pool.Put(data)
			}
		}()
	}

	select {}
}

// GetConnCount get the number of connections in the connection database
func (u *UDPTracker) GetConnCount() int {
	if u == nil || u.conndb == nil {
		return -1
	}
	return u.conndb.conns()
}

// WriteConns writes the connection database to file
func (u *UDPTracker) WriteConns() {
	if u == nil || u.conndb == nil {
		return
	}
	u.conndb.write()
}

func (u *UDPTracker) process(data []byte, remote *net.UDPAddr) {
	var addr [4]byte
	ip := remote.IP.To4()
	copy(addr[:], ip)

	action := data[11]
	txid := int32(binary.BigEndian.Uint32(data[12:16]))

	if ip == nil {
		msg := u.newClientError("IPv6?", txid, cerrFields{"ip": remote.IP.String()})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	if action > 2 {
		msg := u.newClientError("bad action", txid, cerrFields{"action": data[11], "addr": addr})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	if action == 0 {
		c := connect{}
		if err := c.unmarshall(data); err != nil {
			msg := u.newServerError("base.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
		}
		u.connect(&c, remote, addr)
		return
	}

	connid := int64(binary.BigEndian.Uint64(data[0:8]))
	if ok := u.conndb.check(connid, addr); !ok && u.conf.Tracker.UDP.CheckConnID {
		msg := u.newClientError("bad connid", txid, cerrFields{"clientID": connid, "ip": ip})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	switch action {
	case 1:
		if len(data) < 98 {
			msg := u.newClientError("bad announce size", txid, cerrFields{"size": len(data)})
			u.sock.WriteToUDP(msg, remote)
			return
		}
		announce := announce{}
		if err := announce.unmarshall(data); err != nil {
			msg := u.newServerError("announce.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
			return
		}
		u.announce(&announce, remote, addr)
	case 2:
		scrape := scrape{}
		if err := scrape.unmarshall(data); err != nil {
			msg := u.newServerError("scrape.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
			return
		}
		u.scrape(&scrape, remote)
	}
}
