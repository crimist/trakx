/*
	Contains UDP tracker for trakx.
*/

package udp

import (
	"encoding/binary"
	"net"
	"sync"
	"unsafe"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp/protocol"
	"github.com/crimist/trakx/tracker/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	errClosed      = "use of closed network connection"
	requestSizeMax = 1496 // 1496 is max size of a scrape with 20 hashes
)

type UDPTracker struct {
	sock     *net.UDPConn
	conndb   *connectionDatabase
	peerdb   storage.Database
	shutdown chan struct{}
}

// Init sets up the UDPTracker.
func (u *UDPTracker) Init(peerdb storage.Database) {
	u.conndb = newConnectionDatabase(config.Conf.Database.Conn.Timeout)
	u.peerdb = peerdb
	u.shutdown = make(chan struct{})

	go utils.RunOn(config.Conf.Database.Conn.Trim, u.conndb.trim)
}

// Serve begins listening and serving clients.
func (u *UDPTracker) Serve() error {
	var err error

	u.sock, err = net.ListenUDP("udp4", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: config.Conf.Tracker.UDP.Port, Zone: ""})
	if err != nil {
		return errors.Wrap(err, "Failed to open UDP listen socket")
	}

	pool := sync.Pool{
		New: func() interface{} {
			slice := make([]byte, requestSizeMax)
			return &slice
		},
	}

	for i := 0; i < config.Conf.Tracker.UDP.Threads; i++ {
		go func() {
			for {
				data := pool.Get().(*[]byte)
				size, remoteAddr, err := u.sock.ReadFromUDP(*data)
				if err != nil {
					// if socket is closed exit loop
					if errors.Unwrap(err).Error() == errClosed {
						break
					}

					config.Logger.Error("Failed to read from UDP socket", zap.Error(err))
					pool.Put(data)
					continue
				}

				if size > 15 { // 16 = minimum connect
					u.process((*data)[:size], remoteAddr)
				}

				pool.Put(data)
			}
		}()
	}

	<-u.shutdown
	config.Logger.Info("Closing UDP tracker socket")
	if err = u.sock.Close(); err != nil {
		return errors.Wrap(err, "Failed to close UDP listen socket")
	}

	return nil
}

// Shutdown stops the UDP tracker server by closing the socket.
func (u *UDPTracker) Shutdown() {
	if u == nil || u.shutdown == nil {
		return
	}
	var die struct{}
	u.shutdown <- die
}

// ConnCount returns the number of BitTorrent UDP protocol connections in the connection database.
func (u *UDPTracker) ConnCount() int {
	if u == nil || u.conndb == nil {
		return -1
	}
	return u.conndb.conns()
}

// WriteConns writes the connection database to the disk.
func (u *UDPTracker) WriteConns() error {
	if u == nil || u.conndb == nil {
		return nil
	}

	if err := u.conndb.write(); err != nil {
		return errors.Wrap(err, "Failed to write connections database to disk")
	}

	return nil
}

func (u *UDPTracker) process(data []byte, remote *net.UDPAddr) {
	storage.Expvar.Hits.Add(1)
	var cAddr connAddr
	ip := remote.IP.To4()

	copy(cAddr[0:4], ip)
	binary.LittleEndian.PutUint16(cAddr[4:6], uint16(remote.Port))
	addr := *(*[4]byte)(unsafe.Pointer(&cAddr))

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
		c := protocol.Connect{}
		if err := c.Unmarshall(data); err != nil {
			msg := u.newServerError("base.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
		}
		u.connect(&c, remote, cAddr)
		return
	}

	connid := int64(binary.BigEndian.Uint64(data[0:8]))
	if ok := u.conndb.check(connid, cAddr); !ok && config.Conf.Debug.CheckConnIDs {
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
		announce := protocol.Announce{}
		if err := announce.Unmarshall(data); err != nil {
			msg := u.newServerError("announce.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
			return
		}
		u.announce(&announce, remote, addr)
	case 2:
		scrape := protocol.Scrape{}
		if err := scrape.Unmarshall(data); err != nil {
			msg := u.newServerError("scrape.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
			return
		}
		u.scrape(&scrape, remote)
	}
}
