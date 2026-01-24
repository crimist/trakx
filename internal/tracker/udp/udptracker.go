/*
	Contains UDP tracker for trakx.
*/

package udp

import (
	"encoding/binary"
	"net"
	"net/netip"
	"sync"

	"github.com/crimist/trakx/internal/stats"
	"github.com/crimist/trakx/internal/storage"
	"github.com/crimist/trakx/internal/tracker"
	"github.com/crimist/trakx/internal/tracker/udp/connections"
	"github.com/crimist/trakx/internal/tracker/udp/udpprotocol"
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
	peerDB              storage.Database
	config              tracker.TrackerConfig
	collector           stats.Collector
	shutdown            chan struct{}
	sockets             []*net.UDPConn
	connections         *connections.Connections
	validateConnections bool
}

func NewTracker(peerDB storage.Database, config tracker.TrackerConfig, collector stats.Collector, connections *connections.Connections, validateConnections bool) *Tracker {
	return &Tracker{
		peerDB:              peerDB,
		config:              config,
		collector:           collector,
		shutdown:            make(chan struct{}),
		connections:         connections,
		validateConnections: validateConnections,
	}
}

// Serve begins listening and serving clients.
func (tracker *Tracker) Serve(ip net.IP, port int, routines int) error {
	addr := &net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	// Create N sockets with SO_REUSEPORT (one per worker)
	tracker.sockets = make([]*net.UDPConn, routines)
	for i := 0; i < routines; i++ {
		socket, err := listenReusePort("udp", addr)
		if err != nil {
			// Clean up any sockets already created
			for j := 0; j < i; j++ {
				tracker.sockets[j].Close()
			}
			return errors.Wrap(err, "Failed to open UDP listen socket with SO_REUSEPORT")
		}
		tracker.sockets[i] = socket
	}
	zap.L().Info("Serving UDP tracker", zap.String("address", addr.String()), zap.Int("sockets", routines))

	var wg sync.WaitGroup
	wg.Add(routines)

	// Each worker gets its own socket (1:1 mapping)
	for i := 0; i < routines; i++ {
		go func(socket *net.UDPConn) {
			defer wg.Done()
			data := make([]byte, maximumRequestSize)

			for {
				data = data[:cap(data)]
				size, remoteAddr, err := socket.ReadFromUDP(data)
				if err != nil {
					if errors.Unwrap(err).Error() == errSocketClosed {
						break
					}

					zap.L().Error("Failed to read from UDP socket", zap.Error(err))
					continue
				}

				if size < minimumRequestSize {
					socket.WriteToUDP(fatalMinReqLen, remoteAddr)
					zap.L().Debug("client sent packet below minimum request size", zap.String("addr", remoteAddr.String()), zap.Int("size", size), zap.ByteString("data", (data)[:size]))
				} else {
					data = data[:size]
					tracker.process(data, remoteAddr, socket)
				}
			}
		}(tracker.sockets[i])
	}

	<-tracker.shutdown
	zap.L().Info("UDP tracker received shutdown")

	// Close all sockets
	for _, socket := range tracker.sockets {
		if err := socket.Close(); err != nil {
			zap.L().Error("Failed to close UDP tracker socket", zap.Error(err))
		}
	}

	wg.Wait()
	return nil
}

// Shutdown stops the UDP tracker server by closing the socket.
func (tracker *Tracker) Shutdown() {
	var signal struct{}
	tracker.shutdown <- signal
}

func (tracker *Tracker) process(data []byte, udpAddr *net.UDPAddr, socket *net.UDPConn) {
	tracker.collector.Hit()

	action := udpprotocol.Action(data[11])
	transactionID := int32(binary.BigEndian.Uint32(data[12:16]))

	addr, ok := netip.AddrFromSlice(udpAddr.IP)
	if !ok {
		tracker.error(udpAddr, []byte("failed to parse ip"), transactionID, socket)
		zap.L().DPanic("failed to parse remote ip slice as netip", zap.ByteString("ip", udpAddr.IP))
		return
	}
	addr = addr.Unmap() // use ipv4 instead of ipv6 mapped ipv4
	addrPort := netip.AddrPortFrom(addr, uint16(udpAddr.Port))

	if !action.Valid() {
		tracker.error(udpAddr, fatalInvalidAction, transactionID, socket)
		zap.L().Debug("client set invalid action", zap.Binary("packet", data), zap.Uint8("action", data[11]), zap.Any("remote", addrPort))
		return
	}

	switch action {
	case udpprotocol.ActionHeartbeat:
		socket.WriteToUDP(udpprotocol.HeartbeatOk, udpAddr)
		return
	case udpprotocol.ActionConnect:
		tracker.connect(udpAddr, addrPort, transactionID, data, socket)
		return
	}

	connectionID := binary.BigEndian.Uint64(data[0:8])
	if tracker.validateConnections {
		if validConnectionID := tracker.connections.Validate(addrPort, connectionID); !validConnectionID {
			tracker.error(udpAddr, fatalUnregisteredConnection, transactionID, socket)
			zap.L().Debug("client sent unregistered connection id", zap.Binary("packet", data), zap.Uint64("connectionID", connectionID), zap.Any("remote", addrPort))
			return
		}
	} else {
		zap.L().Debug("Skipping UDP connection id validation")
	}

	switch action {
	case udpprotocol.ActionAnnounce:
		tracker.announce(udpAddr, addrPort, transactionID, data, socket)
	case udpprotocol.ActionScrape:
		tracker.scrape(udpAddr, addrPort, transactionID, data, socket)
	}
}
