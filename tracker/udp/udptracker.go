/*
	Contains UDP tracker for trakx.
*/

package udp

import (
	"encoding/binary"
	"net"
	"net/netip"
	"sync"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/stats"
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
	u.conndb = newConnectionDatabase(config.Config.UDP.ConnDB.Expiry)
	u.peerdb = peerdb
	u.shutdown = make(chan struct{})

	if err := u.conndb.loadFromFile(config.CachePath + "conn.db"); err != nil {
		config.Logger.Warn("Failed to load connection database, creating empty db", zap.Error(err))
		u.conndb.make()
	}

	go utils.RunOn(config.Config.UDP.ConnDB.Trim, u.conndb.trim)
}

// Serve begins listening and serving clients.
func (u *UDPTracker) Serve() error {
	var err error

	u.sock, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP(config.Config.UDP.IP),
		Port: config.Config.UDP.Port,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to open UDP listen socket")
	}

	pool := sync.Pool{
		New: func() interface{} {
			slice := make([]byte, requestSizeMax)
			return &slice
		},
	}

	for i := 0; i < config.Config.UDP.Threads; i++ {
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

// Connections returns the number of BitTorrent UDP protocol connections in the connection database.
func (u *UDPTracker) Connections() int {
	if u == nil || u.conndb == nil {
		return -1
	}
	return u.conndb.size()
}

// WriteConns writes the connection database to the disk.
func (u *UDPTracker) WriteConns() error {
	if u == nil || u.conndb == nil {
		return nil
	}

	if err := u.conndb.writeToFile(config.CachePath + "conn.db"); err != nil {
		return errors.Wrap(err, "Failed to write connections database to disk")
	}

	return nil
}

func (u *UDPTracker) process(data []byte, remote *net.UDPAddr) {
	stats.Hits.Add(1)

	action := protocol.Action(data[11])
	txid := int32(binary.BigEndian.Uint32(data[12:16]))

	addr, ok := netip.AddrFromSlice(remote.IP)
	if !ok {
		u.newServerError("failed to parse ip", errors.New("failed to parse remote ip slice as netip"), txid)
	}
	addr = addr.Unmap() // use ipv4 instead of ipv6 mapped ipv4
	addrPort := netip.AddrPortFrom(addr, uint16(remote.Port))

	if action > protocol.ActionHeartbeat {
		msg := u.newClientError("bad action", txid, cerrFields{"action": data[11], "addrPort": addrPort})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	if action == protocol.ActionHeartbeat {
		u.sock.WriteToUDP(protocol.HeartbeatOk, remote)
		return
	}

	if action == protocol.ActionConnect {
		c := protocol.Connect{}
		if err := c.Unmarshall(data); err != nil {
			msg := u.newServerError("base.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
		}
		u.connect(&c, remote, addrPort)
		return
	}

	connid := int64(binary.BigEndian.Uint64(data[0:8]))
	if ok := u.conndb.check(connid, addrPort); !ok && config.Config.UDP.ConnDB.Validate {
		msg := u.newClientError("bad connection id", txid, cerrFields{"clientID": connid, "addrPort": addrPort})
		u.sock.WriteToUDP(msg, remote)
		return
	}

	switch action {
	case protocol.ActionAnnounce:
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

		u.announce(&announce, remote, addrPort)
	case protocol.ActionScrape:
		scrape := protocol.Scrape{}
		if err := scrape.Unmarshall(data); err != nil {
			msg := u.newServerError("scrape.unmarshall()", err, txid)
			u.sock.WriteToUDP(msg, remote)
			return
		}

		u.scrape(&scrape, remote)
	}
}
