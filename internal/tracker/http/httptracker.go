package http

import (
	"bytes"
	"expvar"
	"net"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/crimist/trakx/internal/stats"
	"github.com/crimist/trakx/internal/storage"
	"github.com/crimist/trakx/internal/tracker"
	"github.com/crimist/trakx/internal/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	maximumRequestSize = 2600 // enough for scrapes up to 40 info_hashes
)

type Tracker struct {
	peerdb       storage.Database
	config       tracker.TrackerConfig
	collector    stats.Collector
	shutdown     chan struct{}
	servePath    string
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewTracker(peerDB storage.Database, config tracker.TrackerConfig, collector stats.Collector, servePath string, readTimeout, writeTimeout time.Duration) *Tracker {
	return &Tracker{
		config:       config,
		peerdb:       peerDB,
		shutdown:     make(chan struct{}),
		collector:    collector,
		servePath:    servePath,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

// Serve begins listening and serving clients.
func (tracker *Tracker) Serve(ip net.IP, port int, routines int) error {
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   ip,
		Port: port,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to open TCP listen socket")
	}
	zap.L().Info("Serving HTTP tracker", zap.String("address", listener.Addr().String()), zap.Int("routines", routines))

	// TODO: figure out what optimal number of goroutines is (benchmark)
	// Going to need to write a tool that can simulate a large number of clients
	for i := 0; i < routines; i++ {
		go func() {
			data := make([]byte, maximumRequestSize)

			for {
				data = data[:cap(data)]
				conn, err := listener.Accept()
				if err != nil {
					if errors.Unwrap(err) == net.ErrClosed {
						break
					}

					zap.L().Error("http connection accept failed", zap.Error(err))
					continue
				}

				now := time.Now()
				conn.SetReadDeadline(now.Add(tracker.readTimeout))
				conn.SetWriteDeadline(now.Add(tracker.writeTimeout))

				size, err := conn.Read(data)
				if err != nil {
					zap.L().Error("Failed to read from TCP socket", zap.Error(err))
					continue
				}

				tracker.collector.Hit()
				data = data[:size]
				tracker.process(conn, data)
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

func (tracker *Tracker) process(conn net.Conn, data []byte) {
	defer conn.Close()

	reqData, err := parse(data)
	if errors.Is(err, invalidParse) || reqData.Method != "GET" {
		writeStatus(conn, "400")
		zap.L().Debug("invalid request", zap.Error(err), zap.ByteString("request data", data))
		return
	} else if err != nil {
		zap.L().Error("error parsing request", zap.Error(err), zap.ByteString("request data", data))
		writeStatus(conn, "500")
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
				tracker.error(conn, "Failed to parse X-Forwarded-For")
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
					tracker.error(conn, "Failed to parse remote address")
					break
				}
			}
		}

		var ip netip.Addr
		ip, err := netip.ParseAddr(ipString)
		if err != nil {
			zap.L().Error("Failed to ParseAddr, is X-Forwarded-For enabled?", zap.String("ip", ipString), zap.Error(err))
			tracker.error(conn, "Failed to parse IP address")
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
			tracker.error(conn, "scrape requires infohashes")
			break
		}
		tracker.scrape(conn, reqData.Parameters)
	case "/heartbeat":
		writeStatus(conn, "200")
	case "/stats":
		conn.Write(expvarHeader)
		expvar.Handler().ServeHTTP(ExpvarResponseWriter{
			conn: conn,
		}, nil)
	default:
		if tracker.servePath == "" {
			writeStatus(conn, "404")
			break
		}

		cleanPath := filepath.FromSlash(path.Clean("/" + reqData.Path))
		filePath := filepath.Join(tracker.servePath, cleanPath)

		if !strings.HasPrefix(filePath, tracker.servePath) {
			writeStatus(conn, "403")
			break
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				writeStatus(conn, "404")
			} else {
				zap.L().Debug("Error accessing file", zap.Error(err), zap.String("path", filePath))
				writeStatus(conn, "500")
			}
			break
		}

		if fileInfo.IsDir() {
			writeStatus(conn, "403")
			break
		}

		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			zap.L().Debug("Error reading file", zap.Error(err), zap.String("path", filePath))
			writeStatus(conn, "500")
			break
		}

		contentType := "application/octet-stream"
		switch filepath.Ext(filePath) {
		case ".html", ".htm":
			contentType = "text/html"
		case ".css":
			contentType = "text/css"
		case ".js":
			contentType = "application/javascript"
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".txt":
			contentType = "text/plain"
		}

		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: " + contentType + "\r\nContent-Length: " + strconv.Itoa(len(fileContent)) + "\r\n\r\n" + string(fileContent)))
	}
}
