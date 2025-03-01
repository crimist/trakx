package http

import (
	"bytes"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker"
	"github.com/crimist/trakx/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	maximumRequestSize = 2600 // enough for scrapes up to 40 info_hashes
)

type Tracker struct {
	config        tracker.TrackerConfig
	peerdb        storage.Database
	shutdown      chan struct{}
	stats         *stats.Statistics
	expvarHandler http.Handler
	embeddedCache config.EmbeddedCache
}

func NewTracker(peerDB storage.Database, stats *stats.Statistics, config tracker.TrackerConfig) *Tracker {
	return &Tracker{
		config:        config,
		peerdb:        peerDB,
		shutdown:      make(chan struct{}),
		stats:         stats,
		expvarHandler: expvar.Handler(),
	}
}

// Serve begins listening and serving clients.
func (tracker *Tracker) Serve(ip net.IP, port int, routines int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip.String(), port))
	if err != nil {
		return errors.Wrap(err, "Failed to open TCP listen socket")
	}

	tracker.embeddedCache, err = config.GenerateEmbeddedCache()
	if err != nil {
		return errors.Wrap(err, "failed to generate embedded cache")
	}

	// TODO: figure out what optimal number of goroutines is (benchmark)
	// Going to need to write a tool that can simulate a large number of clients
	for i := 0; i < routines; i++ {
		go func() {
			data := make([]byte, maximumRequestSize)

			for {
				conn, err := listener.Accept()
				if err != nil {
					if errors.Unwrap(err) == net.ErrClosed {
						break
					}

					zap.L().Error("http connection accept failed", zap.Error(err))
					tracker.stats.ServerErrors.Add(1)
					continue
				}

				now := time.Now()
				conn.SetReadDeadline(now.Add(tracker.config.ReadTimeout))
				conn.SetWriteDeadline(now.Add(tracker.config.WriteTimeout))

				size, err := conn.Read(data)
				if err != nil {
					zap.L().Error("Failed to read from TCP socket", zap.Error(err))
					continue
				}

				tracker.stats.Hits.Add(1)
				tracker.process(conn, data, size)
			}
		}()
	}

	<-tracker.shutdown
	if err := listener.Close(); err != nil {
		return errors.Wrap(err, "Failed to close tcp listen socket")
	}

	return nil
}

// Shutdown stops the HTTP tracker server by closing the socket.
func (t *Tracker) Shutdown() {
	if t == nil || t.shutdown == nil {
		return
	}
	var die struct{}
	t.shutdown <- die
}

func (tracker *Tracker) process(conn net.Conn, data []byte, size int) {
	defer conn.Close()

	reqData, err := parse(data, size)
	if err == invalidParse || reqData.Method != "GET" {
		writeStatus(conn, "400")
		return
	} else if err != nil {
		zap.L().Error("error parsing request", zap.Error(err), zap.ByteString("request data", data))
		writeStatus(conn, "500")
		tracker.stats.ServerErrors.Add(1)
		return
	}

	switch reqData.Path {
	case "/announce":
		var params announceParameters
		for _, param := range reqData.Parameters {
			var key, val string

			if equal := bytes.Index(param, []byte("=")); equal == -1 {
				key = string(param) // noescape
				val = "1"
			} else {
				key = string(param[:equal])   // noescape
				val = string(param[equal+1:]) // escape
			}

			switch key {
			case "compact":
				if val == "1" {
					params.compact = true
				}
			case "no_peer_id":
				if val == "1" {
					params.nopeerid = true
				}
			case "left":
				if val == "0" {
					params.noneleft = true
				}
			case "event":
				params.event = val
			case "port":
				params.port = val
			case "info_hash":
				params.hash = val
			case "peer_id":
				params.peerid = val
			case "numwant":
				params.numwant = val
			}
		}

		var ipString string
		forwarded, forwardedIP := parseForwarded(data)
		if forwarded {
			if forwardedIP == nil {
				writeFailure(conn, "Failed to parse X-Forwarded-For")
				tracker.stats.ClientErrors.Add(1)
				break
			}
			ipString = utils.BytesToStringUnsafe(forwardedIP)
		} else {
			if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
				ipString = tcpAddr.IP.String()
			} else {
				zap.L().Debug("non TCPAddr remote address", zap.String("remote addr", conn.RemoteAddr().String()))
				ipString, _, err = net.SplitHostPort(conn.RemoteAddr().String())
				if err != nil {
					zap.L().Error("Failed to SplitHostPort", zap.Error(err))
					writeFailure(conn, "Failed to parse remote address")
					tracker.stats.ClientErrors.Add(1)
					break
				}
			}
		}

		var ip netip.Addr
		ip, err := netip.ParseAddr(ipString)
		if err != nil {
			zap.L().Error("Failed to ParseAddr, is X-Forwarded-For enabled?", zap.String("ip", ipString), zap.Error(err))
			writeFailure(conn, "Failed to parse IP address")
			tracker.stats.ClientErrors.Add(1)
			break
		}

		tracker.announce(conn, &params, ip)
	case "/scrape":
		var count int
		for i := 0; i < len(reqData.Parameters); i++ {
			if len(reqData.Parameters[i]) < 10 || !bytes.Equal(reqData.Parameters[i][0:10], []byte("info_hash=")) {
				reqData.Parameters[i] = nil
			} else {
				reqData.Parameters[i] = reqData.Parameters[i][10:]
				count++
			}
		}
		if count == 0 {
			writeFailure(conn, "scrape requires infohashes")
			tracker.stats.ClientErrors.Add(1)
			break
		}
		tracker.scrape(conn, reqData.Parameters)
	case "/heartbeat":
		writeStatus(conn, "200")
	case "/stats":
		conn.Write(expvarHeader)
		tracker.expvarHandler.ServeHTTP(ExpvarResponseWriter{
			conn: conn,
		}, nil)
	default:
		if data, ok := tracker.embeddedCache[reqData.Path]; ok {
			writeSuccess(conn, data)
		} else {
			writeStatus(conn, "404")
		}
	}

	conn.Close()
}
