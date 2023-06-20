/*
	Contains UDP tracker for trakx.
*/

// TODO: rename this to `udp`
package udptracker

import (
	"encoding/binary"
	"net"
	"net/netip"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker/udptracker/conncache"
	"github.com/crimist/trakx/tracker/udptracker/udpprotocol"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	errSocketClosed = "use of closed network connection"
	// maximum request size (a scrape request with 20/20 hashes)
	maximumRequestSize = 1496
	// min request size (a connect request)
	minimumRequestSize  = 16
	minimumAnnounceSize = 98
)

var (
	// need to keep it short to prevent udp amplification
	requestTooSmall = []byte("nope")
)

type TrackerConfig struct {
	Validate         bool
	DefaultNumwant   int
	MaximumNumwant   int
	Interval         int32
	IntervalVariance int32
}

type Tracker struct {
	config    TrackerConfig
	socket    *net.UDPConn
	connCache *conncache.ConnectionCache
	peerDB    storage.Database
	shutdown  chan struct{}
	stats     *stats.Statistics
}

func NewTracker(peerDB storage.Database, connCache *conncache.ConnectionCache, stats *stats.Statistics, config TrackerConfig) *Tracker {
	tracker := Tracker{
		config:    config,
		peerDB:    peerDB,
		connCache: connCache,
		shutdown:  make(chan struct{}),
		stats:     stats,
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

	requestPool := pools.NewPool(func() any {
		return make([]byte, maximumRequestSize)
	}, func(slice []byte) {
		slice = slice[:cap(slice)]
	})

	// TODO: figure out what optimal number of goroutines is (benchmark)
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
	return tracker.connCache.EntryCount()
}

func (tracker *Tracker) process(data []byte, udpAddr *net.UDPAddr) {
	if tracker.stats != nil {
		tracker.stats.Hits.Add(1)
	}

	action := udpprotocol.Action(data[11])
	transactionID := int32(binary.BigEndian.Uint32(data[12:16]))

	addr, ok := netip.AddrFromSlice(udpAddr.IP)
	if !ok {
		tracker.sendError(udpAddr, "failed to parse ip", transactionID)
		zap.L().DPanic("failed to parse remote ip slice as netip", zap.ByteString("ip", udpAddr.IP))
		return
	}
	addr = addr.Unmap() // use ipv4 instead of ipv6 mapped ipv4
	addrPort := netip.AddrPortFrom(addr, uint16(udpAddr.Port))

	if action.IsInvalid() {
		tracker.sendError(udpAddr, "invalid action", transactionID)
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
		if validConnectionID := tracker.connCache.Validate(connectionID, addrPort); !validConnectionID {
			tracker.sendError(udpAddr, "unregistered connection id", transactionID)
			zap.L().Debug("client sent unregistered connection id", zap.Binary("packet", data), zap.Int64("connectionID", connectionID), zap.Any("remote", addrPort))
			return
		}
	} else {
		zap.L().Debug("insecure beaviour - skipping connection id validation")
	}

	switch action {
	case udpprotocol.ActionAnnounce:
		tracker.announce(udpAddr, addrPort, transactionID, data)
	case udpprotocol.ActionScrape:
		tracker.scrape(udpAddr, addrPort, transactionID, data)
	}
}
