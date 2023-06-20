package http

import (
	"bytes"
	"expvar"
	"net"
	"net/netip"
	"time"
	"unsafe"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/tracker/stats"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type workers struct {
	tracker   *HTTPTracker
	listener  net.Listener
	fileCache config.EmbeddedCache
}

func (w *workers) startWorkers(num int) {
	config.Logger.Debug("Starting http workers", zap.Int("count", num))
	for i := 0; i < num; i++ {
		go w.work()
	}
}

func (w *workers) work() {
	expvarHandler := expvar.Handler()
	statRespWriter := fakeRespWriter{}
	data := make([]byte, httpRequestMax)

	for {
		conn, err := w.listener.Accept()
		if err != nil {
			// if socket is closed we're done
			if errors.Unwrap(err) == net.ErrClosed {
				break
			}

			// otherwise log the error
			config.Logger.Error("http connection accept failed", zap.Error(err))
			stats.ServerErrors.Add(1)
			continue
		}

		now := time.Now()
		conn.SetReadDeadline(now.Add(config.Config.HTTP.Timeout.Read))
		conn.SetWriteDeadline(now.Add(config.Config.HTTP.Timeout.Write))

		size, err := conn.Read(data)
		if err != nil {
			conn.Close()
			continue
		}
		stats.Hits.Add(1)

		p, err := parse(data, size)
		if err == invalidParse || p.Method != "GET" {
			// invalid request
			writeStatus(conn, "400")
			conn.Close()
			continue
		} else if err != nil {
			// error in parse
			config.Logger.Error("error parsing request", zap.Error(err), zap.Any("request data", data))
			writeStatus(conn, "500")
			conn.Close()

			stats.ServerErrors.Add(1)
			continue
		}

		switch p.Path {
		case "/announce":
			var v announceParams
			for _, param := range p.Params {
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
						v.compact = true
					}
				case "no_peer_id":
					if val == "1" {
						v.nopeerid = true
					}
				case "left":
					if val == "0" {
						v.noneleft = true
					}
				case "event":
					v.event = val
				case "port":
					v.port = val
				case "info_hash":
					v.hash = val
				case "peer_id":
					v.peerid = val
				case "numwant":
					v.numwant = val
				}
			}

			var ip netip.Addr
			var ipStr string

			forwarded, forwardedIP := parseForwarded(data)
			if forwarded {
				if forwardedIP == nil {
					w.tracker.clientError(conn, "Failed to parse X-Forwarded-For")
					break
				}
				ipStr = *(*string)(unsafe.Pointer(&forwardedIP))
			} else {
				ipStr, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
			}

			ip, err := netip.ParseAddr(ipStr)
			if err != nil {
				config.Logger.Warn("Failed to parse value from X-Forwarded-For", zap.String("ip string", ipStr), zap.Error(err))
				w.tracker.clientError(conn, "Failed to parse forwarded IP")
				break
			}

			w.tracker.announce(conn, &v, ip)
		case "/scrape":
			var count int
			for i := 0; i < len(p.Params); i++ {
				if len(p.Params[i]) < 10 || !bytes.Equal(p.Params[i][0:10], []byte("info_hash=")) {
					p.Params[i] = nil
				} else {
					p.Params[i] = p.Params[i][10:]
					count++
				}
			}
			if count == 0 {
				w.tracker.clientError(conn, "no infohashes")
				break
			}
			w.tracker.scrape(conn, p.Params)
		case "/heartbeat":
			writeStatus(conn, "200")
		case "/stats":
			// Serves expvar handler but it's hacky af
			statRespWriter.conn = conn

			conn.Write(statsHeader)
			expvarHandler.ServeHTTP(statRespWriter, nil)
		default:
			// check if file is embedded
			if data, ok := w.fileCache[p.Path]; ok {
				writeData(conn, data)
			} else {
				// otherwise return 404
				writeStatus(conn, "404")
			}
		}

		conn.Close()
	}
}
