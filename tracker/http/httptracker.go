package http

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"runtime"
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
	t.workers.pool.New = func() interface{} { return make([]byte, 1000, 1000) } // TODO: HTTP req max size?
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
	maxread := time.Duration(w.tracker.conf.Tracker.HTTP.ReadTimeout) * time.Second
	maxwrite := time.Duration(w.tracker.conf.Tracker.HTTP.WriteTimeout) * time.Second
	for {
		select {
		case job := <-w.jobQueue:
			func() {
				data := w.pool.Get().([]byte)
				defer func() {
					job.conn.Close()
					for i := range data {
						data[i] = 0
					}
					w.pool.Put(data)
				}()

				// Should recv and send data within timeouts or were overloaded
				now := time.Now()
				job.conn.SetDeadline(now.Add(maxread))
				job.conn.SetWriteDeadline(now.Add(maxwrite))

				if _, err := job.conn.Read(data); err != nil {
					return
				}

				urlEnd := bytes.Index(data, []byte(" HTTP/"))
				if urlEnd < 5 {
					job.conn.Write([]byte("HTTP/1.1 400\r\n\r\n"))
					return
				}

				u, err := url.Parse(string(data[4:urlEnd]))
				if err != nil {
					job.conn.Write([]byte("HTTP/1.1 400\r\n\r\n"))
					return
				}

				switch u.Path {
				case "/announce":
					w.tracker.announce(job.conn, u.Query())
				case "/scrape":
					w.tracker.scrape(job.conn, u.Query())
				case "/":
					job.writeData(w.index)
				case "/dmca":
					job.redir("https://www.youtube.com/watch?v=BwSts2s4ba4")
				case "/stats":
					// httptracker.QueueLen()
					// udptracker.GetConnCount()
					job.writeData(fmt.Sprintf("Hashes %d Ann: %d/%dOK Scr: %d/%dOK Con: %d/%dOK Err: %dServer/%dClient/s Goroutines: %d", w.tracker.peerdb.Hashes(), shared.Expvar.Announces, shared.Expvar.AnnouncesOK, shared.Expvar.Scrapes, shared.Expvar.ScrapesOK, shared.Expvar.Connects, shared.Expvar.ConnectsOK, shared.Expvar.Errs, shared.Expvar.Clienterrs, runtime.NumGoroutine()))
				default:
					job.writeStatus("404")
				}
			}()
		}
	}
}
