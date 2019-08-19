package http

import (
	"expvar"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type HTTPTracker struct {
	conf   *shared.Config
	logger *zap.Logger
	peerdb *shared.PeerDatabase

	workers workers
}

func NewHTTPTracker(conf *shared.Config, logger *zap.Logger, peerdb *shared.PeerDatabase) *HTTPTracker {
	tracker := HTTPTracker{
		conf:   conf,
		logger: logger,
		peerdb: peerdb,
	}

	return &tracker
}

func (t *HTTPTracker) Serve(index []byte, threads int) {
	t.workers = workers{
		tracker:  t,
		jobQueue: make(chan job, t.conf.Tracker.HTTP.Qsize),
		index:    string(index),
	}

	// TODO: Find HTTP req max size
	// Note: go 1.13 will finally allow pools to survive past GC cycle so
	// this will be far more efficient as our load is quite consistant
	t.workers.pool.New = func() interface{} { return make([]byte, 1000, 1000) }
	t.workers.startWorkers(threads)

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.conf.Tracker.HTTP.Port))
	if err != nil {
		t.logger.Panic("net.Listen()", zap.Error(err))
	}

	for i := 0; i < t.conf.Tracker.HTTP.Accepters; i++ {
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					if !t.conf.Trakx.Prod {
						t.logger.Warn("net.Listen()", zap.Error(err))
					}
					continue
				}

				// If jobQueue ever locks we should stop accepting packets anyway
				t.workers.jobQueue <- job{conn}
			}
		}()
	}

	select {}
}

func (t *HTTPTracker) QueueLen() int {
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

			l, err := j.conn.Read(data)
			if err != nil {
				break
			}

			p := parse(&data, l)
			if p.URLend < 5 { // less than "GET / HTTP..."
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
				w.tracker.announce(j.conn, &v)
			case "/scrape":
				// TODO: custom parsing
				u, err := url.Parse(string(data[4:p.URLend]))
				if err != nil {
					j.writeStatus("400")
					break
				}
				w.tracker.scrape(j.conn, u.Query())
			case "/":
				j.writeData(w.index)
			case "/dmca":
				j.redir("https://www.youtube.com/watch?v=BwSts2s4ba4")
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
