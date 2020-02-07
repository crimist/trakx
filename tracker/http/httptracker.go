package http

import (
	"errors"
	"expvar"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

const (
	httpRequestMax = 2100 // slight buffer over 2000
	errClosed      = "use of closed network connection"

	DMCAData = `
	<p>tracker@nibba.trade</p>
	<iframe width="560" height="315" src="https://www.youtube.com/embed/BwSts2s4ba4?controls=0&showinfo=0&autoplay=1" frameborder="0" allowfullscreen></iframe>
	<p>Trakx does not have the capability to block hashes nor does it store or distribute any content.</p>
	`
)

type HTTPTracker struct {
	conf    *shared.Config
	logger  *zap.Logger
	peerdb  storage.Database
	workers workers
	kill    chan struct{}
}

func (t *HTTPTracker) Init(conf *shared.Config, logger *zap.Logger, peerdb storage.Database) {
	t.conf = conf
	t.logger = logger
	t.peerdb = peerdb
	t.kill = make(chan struct{})
}

func (t *HTTPTracker) Serve(index []byte) {
	t.workers = workers{
		tracker:  t,
		jobQueue: make(chan job, t.conf.Tracker.HTTP.Qsize),
		index:    string(index),
	}

	t.workers.pool.New = func() interface{} {
		println("New bytes from worker pool")
		return make([]byte, httpRequestMax, httpRequestMax)
	}
	t.workers.startWorkers(t.conf.Tracker.HTTP.Threads)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.conf.Tracker.HTTP.Port))
	if err != nil {
		t.logger.Panic("net.Listen()", zap.Error(err))
	}

	for i := 0; i < t.conf.Tracker.HTTP.Accepters; i++ {
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					if errors.Unwrap(err).Error() == errClosed { // if socket is closed we're done
						break
					}
					t.logger.Warn("net.Listen()", zap.Error(err))
					continue
				}

				// If jobQueue ever locks we should stop accepting packets anyway
				t.workers.jobQueue <- job{conn}
			}
		}()
	}

	select {
	case _ = <-t.kill:
		t.logger.Info("Closing HTTP tracker connection")
		ln.Close()
	}
}

// Kill kills the HTTP tracker by closing the listening connection
func (t *HTTPTracker) Kill() {
	if t == nil || t.kill == nil {
		return
	}
	var die struct{}
	t.kill <- die
}

func (t *HTTPTracker) QueueLen() int {
	if t == nil || t.workers.jobQueue == nil {
		return -1
	}
	return len(t.workers.jobQueue)
}

type job struct {
	conn net.Conn
}

func (j *job) redir(url string) {
	j.conn.Write([]byte("HTTP/1.1 303\r\nLocation: " + url + "\r\n\r\n"))
}

func (j *job) writeData(data string) {
	j.conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + data))
}

func (j *job) writeStatus(status string) {
	j.conn.Write([]byte("HTTP/1.1 " + status + "\r\n\r\n"))
}

type workers struct {
	tracker  *HTTPTracker
	jobQueue chan job
	pool     sync.Pool

	index string
}

func (w *workers) startWorkers(num int) {
	for i := 0; i < num; i++ {
		go w.work()
	}
}

func (w *workers) work() {
	var j job
	expvarHandler := expvar.Handler()
	maxread := time.Duration(w.tracker.conf.Tracker.HTTP.ReadTimeout) * time.Second
	maxwrite := time.Duration(w.tracker.conf.Tracker.HTTP.WriteTimeout) * time.Second

	for {
		data := w.pool.Get().([]byte)
		select {
		case j = <-w.jobQueue:
			// Should recv and send data within timeouts or were overloaded
			now := time.Now()
			j.conn.SetDeadline(now.Add(maxread))
			j.conn.SetWriteDeadline(now.Add(maxwrite))

			_, err := j.conn.Read(data)
			if err != nil {
				break
			}

			p, err := parse(data)
			if p == nil { // invalid request
				j.writeStatus("400")
				break
			}
			if err != nil { // error in parse
				w.tracker.logger.Error("parse()", zap.Error(err), zap.Any("data", data))
				j.writeStatus("500")
				break
			}
			if p.URLend < 5 || p.Method != "GET" { // less than "GET / HTTP..."
				j.writeStatus("400")
				break
			}

			switch p.Path {
			case "/announce":
				var v announceParams
				for _, param := range p.Params {
					var key, val string

					if equal := strings.Index(param, "="); equal == -1 {
						key = param
						val = "1"
					} else {
						key = param[:equal]
						val = param[equal+1:]
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
						w.tracker.clientError(j.conn, "Bad IP, potentially heroku issue")
						break
					}
					ipStr = *(*string)(unsafe.Pointer(&forwardedIP))
				} else {
					// Not appeng
					ipStr, _, _ = net.SplitHostPort(j.conn.RemoteAddr().String())
				}
				parsedIP := net.ParseIP(ipStr).To4()
				if parsedIP == nil {
					w.tracker.clientError(j.conn, "IPv6 unsupported")
					break
				}
				copy(ip[:], parsedIP)

				w.tracker.announce(j.conn, &v, ip)
			case "/scrape":
				// Custom parsing isn't needed since we aren't hammered with scrapes
				u, err := url.Parse(string(data[4:p.URLend]))
				if err != nil {
					j.writeStatus("400")
					break
				}
				w.tracker.scrape(j.conn, u.Query())
			case "/":
				j.writeData(w.index)
			case "/dmca":
				j.writeData(DMCAData)
			case "/stats":
				// Serves expvar handler but it's hacky af
				j.conn.Write([]byte("HTTP/1.1 200\r\nContent-Type: application/json; charset=utf-8\r\n\r\n"))
				expvarHandler.ServeHTTP(&fakeRespWriter{conn: j.conn}, nil)
			default:
				j.writeStatus("404")
			}
		}

		w.pool.Put(data)
		j.conn.Close()
	}
}
