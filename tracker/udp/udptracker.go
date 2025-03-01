/*
	Contains UDP tracker for trakx.
*/

package udp

import (
	"encoding/binary"
	"net"
	"net/netip"

	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker"
	"github.com/crimist/trakx/tracker/udp/connections"
	"github.com/crimist/trakx/tracker/udp/udpprotocol"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	errSocketClosed = "use of closed network connection"
	// maximum request size, derived from the size of a scrape with 74 hashes
	maximumRequestSize = 16 + 20*74
	// minimum request size, derived from the size of a connect request
	minimumRequestSize  = 16
	minimumAnnounceSize = 98
)

var (
	// fatalMinReqLen needs to be short to prevent UDP amplication abuse
	fatalMinReqLen              = []byte("nope")
	fatalInvalidAction          = []byte("invalid action")
	fatalUnregisteredConnection = []byte("unregistered connection id")
)

type Tracker struct {
	config      tracker.TrackerConfig
	socket      *net.UDPConn
	connections *connections.Connections
	peerDB      storage.Database
	shutdown    chan struct{}
	stats       *stats.Statistics
}

func NewTracker(peerDB storage.Database, connections *connections.Connections, stats *stats.Statistics, config tracker.TrackerConfig) *Tracker {
	return &Tracker{
		config:      config,
		peerDB:      peerDB,
		connections: connections,
		shutdown:    make(chan struct{}),
		stats:       stats,
	}
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

	// TODO: figure out what optimal number of goroutines is (benchmark)
	// Going to need to write a tool that can simulate a large number of clients
	for i := 0; i < routines; i++ {
		go func() {
			data := make([]byte, maximumRequestSize)
			for {
				size, remoteAddr, err := tracker.socket.ReadFromUDP(data)
				if err != nil {
					if errors.Unwrap(err).Error() == errSocketClosed {
						break
					}

					zap.L().Error("Failed to read from UDP socket", zap.Error(err))
					continue
				}

				if size < minimumRequestSize {
					tracker.socket.WriteToUDP(fatalMinReqLen, remoteAddr)
					zap.L().Debug("client sent packet below minimum request size", zap.String("addr", remoteAddr.String()), zap.Int("size", size), zap.ByteString("data", (data)[:size]))
				} else {
					tracker.process((data)[:size], remoteAddr)
				}
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

func (tracker *Tracker) process(data []byte, udpAddr *net.UDPAddr) {
	if tracker.stats != nil {
		tracker.stats.Hits.Add(1)
	}

	action := udpprotocol.Action(data[11])
	transactionID := int32(binary.BigEndian.Uint32(data[12:16]))

	addr, ok := netip.AddrFromSlice(udpAddr.IP)
	if !ok {
		tracker.fatal(udpAddr, []byte("failed to parse ip"), transactionID)
		zap.L().DPanic("failed to parse remote ip slice as netip", zap.ByteString("ip", udpAddr.IP))
		return
	}
	addr = addr.Unmap() // use ipv4 instead of ipv6 mapped ipv4
	addrPort := netip.AddrPortFrom(addr, uint16(udpAddr.Port))

	if !action.Valid() {
		tracker.fatal(udpAddr, fatalInvalidAction, transactionID)
		zap.L().Debug("client set invalid action", zap.Binary("packet", data), zap.Uint8("action", data[11]), zap.Any("remote", addrPort))
		return
	}

	switch action {
	case udpprotocol.ActionHeartbeat:
		tracker.socket.WriteToUDP(udpprotocol.HeartbeatOk, udpAddr)
		return
	case udpprotocol.ActionConnect:
		tracker.connect(udpAddr, addrPort, transactionID, data)
		return
	}

	connectionID := int64(binary.BigEndian.Uint64(data[0:8]))
	if tracker.config.Validate {
		if validConnectionID := tracker.connections.Validate(addrPort, connectionID); !validConnectionID {
			tracker.fatal(udpAddr, fatalUnregisteredConnection, transactionID)
			zap.L().Debug("client sent unregistered connection id", zap.Binary("packet", data), zap.Int64("connectionID", connectionID), zap.Any("remote", addrPort))
			return
		}
	} else {
		zap.L().Debug("Skipping UDP connection id validation")
	}

	switch action {
	case udpprotocol.ActionAnnounce:
		tracker.announce(udpAddr, addrPort, transactionID, data)
	case udpprotocol.ActionScrape:
		tracker.scrape(udpAddr, addrPort, transactionID, data)
	}
}
