/*
	Contains UDP tracker for trakx.
*/

package udptracker

import (
	"encoding/binary"
	"net"
	"net/netip"
	"time"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/crimist/trakx/tracker/udptracker/protocol"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	errSocketClosed = "use of closed network connection"
	// maximum request size (a scrape request with 20/20 hashes)
	maximumRequestSize = 1496
	// min request size (a connect request)
	minimumRequestSize = 16
)

var (
	// need to keep it short to prevent udp amplification
	requestTooSmall = []byte("nope")
)

type TrackerConfig struct {
	Validate                  bool
	peerExpiry                time.Duration
	connDatabasePath          string
	connDatabaseSize          int
	connDatabaseTrimFrequency time.Duration
}

type Tracker struct {
	config   TrackerConfig
	socket   *net.UDPConn
	connDB   *connectionDatabase
	peerDB   storage.Database
	shutdown chan struct{}
}

func NewTracker(peerDB storage.Database, connDB *connectionDatabase, config TrackerConfig) *Tracker {
	tracker := Tracker{
		config:   config,
		peerDB:   peerDB,
		connDB:   connDB,
		shutdown: make(chan struct{}),
	}

	return &tracker
}

// Serve begins listening and serving clients.
func (tracker *Tracker) Serve(ip net.IP, port int, routines int) error {
	var err error

	tracker.socket, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to open UDP listen socket")
	}

	requestPool := pools.NewPool[[]byte](func() any {
		return make([]byte, maximumRequestSize)
	}, func(slice []byte) {
		slice = slice[:cap(slice)]
	})

	// TODO: figure out what optimal number of goroutines is
	for i := 0; i < routines; i++ {
		go func() {
			for {
				data := requestPool.Get()
				size, remoteAddr, err := tracker.socket.ReadFromUDP(data)
				if err != nil {
					if errors.Unwrap(err).Error() == errSocketClosed {
						break
					}

					zap.L().Error("Failed to read from UDP socket", zap.Error(err))
					requestPool.Put(data)
					continue
				}

				if size < minimumRequestSize {
					zap.L().Warn("UDP packet below minimum request size", zap.String("addr", remoteAddr.String()), zap.Int("size", size), zap.ByteString("data", (data)[:size]))
					tracker.socket.WriteToUDP(requestTooSmall, remoteAddr)
				} else {
					// TODO: is there a way not slice the data here?
					tracker.process((data)[:size], remoteAddr)
				}

				requestPool.Put(data)
			}
		}()
	}

	<-tracker.shutdown
	zap.L().Info("UDP trakcer received shutdown")

	if err = tracker.socket.Close(); err != nil {
		return errors.Wrap(err, "Failed to close UDP tracker socket")
	}

	return nil
}

// Shutdown stops the UDP tracker server by closing the socket.
func (tracker *Tracker) Shutdown() {
	var signal struct{}
	tracker.shutdown <- signal
}

// ConnectionCount returns the number of BitTorrent UDP protocol connections in the connection database.
func (tracker *Tracker) ConnectionCount() int {
	if tracker == nil || tracker.connDB == nil {
		return -1
	}
	return tracker.connDB.size()
}

func (tracker *Tracker) process(data []byte, remote *net.UDPAddr) {
	stats.Hits.Add(1)

	action := protocol.Action(data[11])
	transactionID := int32(binary.BigEndian.Uint32(data[12:16]))

	remoteAddr, ok := netip.AddrFromSlice(remote.IP)
	if !ok {
		tracker.sendError(remote, "failed to parse ip", transactionID)
		zap.L().DPanic("failed to parse remote ip slice as netip", zap.ByteString("ip", remote.IP))
		return
	}
	remoteAddr = remoteAddr.Unmap() // use ipv4 instead of ipv6 mapped ipv4
	remoteAddrPort := netip.AddrPortFrom(remoteAddr, uint16(remote.Port))

	if action >= protocol.ActionInvalid {
		tracker.sendError(remote, "invalid action", transactionID)
		zap.L().Debug("client set invalid action", zap.Binary("packet", data), zap.Uint8("action", data[11]), zap.Any("remote", remoteAddrPort))
		return
	}

	if action == protocol.ActionHeartbeat {
		tracker.socket.WriteToUDP(protocol.HeartbeatOk, remote)
		return
	}

	if action == protocol.ActionConnect {
		connect := protocol.Connect{}
		if err := connect.Unmarshall(data); err != nil {
			tracker.sendError(remote, "failed to parse connect request", transactionID)
			zap.L().Debug("client sent invalid connect request", zap.Binary("packet", data), zap.Any("remote", remoteAddrPort))
			return
		}
		tracker.connect(connect, remote, remoteAddrPort)
		return
	}

	// TODO: pick up refactoring from here, done with error.go

	connid := int64(binary.BigEndian.Uint64(data[0:8]))
	if ok := tracker.connDB.check(connid, remoteAddrPort); !ok && tracker.config.Validate {
		msg := tracker.newClientError("bad connection id", transactionID, cerrFields{"clientID": connid, "addrPort": remoteAddrPort})
		tracker.socket.WriteToUDP(msg, remote)
		return
	}

	switch action {
	case protocol.ActionAnnounce:
		if len(data) < 98 {
			msg := tracker.newClientError("bad announce size", transactionID, cerrFields{"size": len(data)})
			tracker.socket.WriteToUDP(msg, remote)
			return
		}

		announce := protocol.Announce{}
		if err := announce.Unmarshall(data); err != nil {
			msg := tracker.newServerError("announce.unmarshall()", err, transactionID)
			tracker.socket.WriteToUDP(msg, remote)
			return
		}

		tracker.announce(&announce, remote, remoteAddrPort)
	case protocol.ActionScrape:
		scrape := protocol.Scrape{}
		if err := scrape.Unmarshall(data); err != nil {
			msg := tracker.newServerError("scrape.unmarshall()", err, transactionID)
			tracker.socket.WriteToUDP(msg, remote)
			return
		}

		tracker.scrape(&scrape, remote)
	}
}
