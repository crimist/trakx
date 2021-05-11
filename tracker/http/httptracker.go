package http

import (
	"bytes"
	"errors"
	"expvar"
	"fmt"
	"net"
	"time"
	"unsafe"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

const (
	httpRequestMax = 2600 // enough for scrapes up to 40 info_hashes
)

var httpSuccess = "HTTP/1.1 200\r\n\r\n"

type HTTPTracker struct {
	conf     *shared.Config
	logger   *zap.Logger
	peerdb   storage.Database
	workers  workers
	shutdown chan struct{}
}

// Init sets the HTTP trackers required values
func (t *HTTPTracker) Init(conf *shared.Config, peerdb storage.Database) {
	t.conf = conf
	t.logger = conf.Logger
	t.peerdb = peerdb
	t.shutdown = make(chan struct{})
}

// Serve starts the HTTP service and begins to serve clients
func (t *HTTPTracker) Serve() {
	t.workers = workers{
		tracker: t,
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.conf.Tracker.HTTP.Port))
	if err != nil {
		t.logger.Panic("net.Listen()", zap.Error(err))
	}

	t.workers.startWorkers(t.conf.Tracker.HTTP.Threads, ln)

	<-t.shutdown
	t.logger.Info("Closing HTTP tracker listen socket")
	ln.Close()
}

// Shutdown gracefully closes the HTTP service by closing the listening connection
func (t *HTTPTracker) Shutdown() {
	if t == nil || t.shutdown == nil {
		return
	}
	var die struct{}
	t.shutdown <- die
}

func redir(c net.Conn, url string) {
	c.Write(shared.StringToBytes("HTTP/1.1 303\r\nLocation: " + url + "\r\n\r\n"))
}

func writeData(c net.Conn, data string) {
	c.Write(shared.StringToBytes("HTTP/1.1 200\r\n\r\n" + data))
}

func writeDataBytes(c net.Conn, data []byte) {
	c.Write(append([]byte("HTTP/1.1 200\r\n\r\n"), data...))
}

func writeStatus(c net.Conn, status string) {
	c.Write(shared.StringToBytes("HTTP/1.1 " + status + "\r\n\r\n"))
}

type workers struct {
	tracker *HTTPTracker
}

func (w *workers) startWorkers(num int, ln net.Listener) {
	w.tracker.logger.Debug("Starting http workers", zap.Int("count", num))
	for i := 0; i < num; i++ {
		go w.work(ln)
	}
}

func (w *workers) work(ln net.Listener) {
	expvarHandler := expvar.Handler()
	statRespWriter := fakeRespWriter{}
	maxread := time.Duration(w.tracker.conf.Tracker.HTTP.ReadTimeout) * time.Second
	maxwrite := time.Duration(w.tracker.conf.Tracker.HTTP.WriteTimeout) * time.Second
	data := make([]byte, httpRequestMax)

	for {
		conn, err := ln.Accept()
		if err != nil {
			// if socket is closed we're done
			if errors.Unwrap(err) == net.ErrClosed {
				break
			}

			// otherwise log the error
			storage.Expvar.Errors.Add(1)
			w.tracker.logger.Warn("net.Listen()", zap.Error(err))
			continue
		}

		now := time.Now()
		conn.SetDeadline(now.Add(maxread))
		conn.SetWriteDeadline(now.Add(maxwrite))

		size, err := conn.Read(data)
		if err != nil {
			conn.Close()
			continue
		}
		storage.Expvar.Hits.Add(1)

		p, err := parse(data, size)
		if err == invalidParse || p.Method != "GET" {
			// invalid request
			writeStatus(conn, "400")
			conn.Close()
			continue
		} else if err != nil {
			// error in parse
			storage.Expvar.Errors.Add(1)
			w.tracker.logger.Error("error parsing request", zap.Error(err), zap.Any("request data", data))
			writeStatus(conn, "500")

			conn.Close()
			continue
		}

		switch p.Path {
		case "/announce":
			var v announceParams
			for _, param := range p.Params {
				var key, val string

				if equal := bytes.Index(param, []byte("=")); equal == -1 {
					key = string(param) // doesn't escape
					val = "1"
				} else {
					key = string(param[:equal])   // doesn't escape
					val = string(param[equal+1:]) // escapes - TODO: optimize?
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

			var ip storage.PeerIP
			var ipStr string

			forwarded, forwardedIP := getForwarded(data)
			if forwarded {
				// Appeng (heroku)
				if forwardedIP == nil {
					w.tracker.clientError(conn, "Bad IP, potentially heroku issue")
					break
				}
				ipStr = *(*string)(unsafe.Pointer(&forwardedIP))
			} else {
				// Not appeng
				ipStr, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
			}

			if err := ip.Set(ipStr); err != nil {
				w.tracker.conf.Logger.Warn("failed to parse ip", zap.String("ip", ipStr), zap.Error(err), zap.Any("attempt", ip))

				w.tracker.clientError(conn, "failed to parse ip: "+err.Error())
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
		case "/":
			writeData(conn, shared.IndexData)
		case "/dmca":
			writeData(conn, shared.DMCAData)
		case "/stats":
			// Serves expvar handler but it's hacky af
			statRespWriter.conn = conn

			conn.Write(shared.StringToBytes("HTTP/1.1 200\r\nContent-Type: application/json; charset=utf-8\r\n\r\n"))
			expvarHandler.ServeHTTP(statRespWriter, nil)
		default:
			writeStatus(conn, "404")
		}

		conn.Close()
	}
}
