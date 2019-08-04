package http

import (
	"bytes"
	"fmt"
	"net"
	"net/url"
	"sync"

	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type HTTPTracker struct {
	conf   *shared.Config
	logger *zap.Logger
	peerdb *shared.PeerDatabase
}

func NewHTTPTracker(conf *shared.Config, logger *zap.Logger, peerdb *shared.PeerDatabase) *HTTPTracker {
	tracker := HTTPTracker{
		conf:   conf,
		logger: logger,
		peerdb: peerdb,
	}

	return &tracker
}

type ctx struct {
	conn net.Conn
	u    *url.URL
}

func (t *HTTPTracker) Serve(index []byte) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", t.conf.Tracker.HTTP.Port))
	if err != nil {
		t.logger.Panic("net.Listen()", zap.Error(err))
	}

	var pool sync.Pool
	pool.New = func() interface{} { return make([]byte, 1000, 1000) } // TODO: HTTP req max size?

	w := workers{
		tracker: t,
		jobQueue: make(chan job, 5000),
	}

	for i := 0; i < 5; i++ {
		go w.consumeAnnounce()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			if !t.conf.Trakx.Prod {
				t.logger.Warn("net.Listen()", zap.Error(err))
			}
			continue
		}
		go func() {
			data := pool.Get().([]byte)
			defer func() {
				for i := range data {
					data[i] = 0
				}
				pool.Put(data)
			}()

			conn.Read(data)

			// Do stuff
			if !bytes.Equal(data[0:4], []byte("GET ")) {
				conn.Write([]byte("Trakx only supports GET"))
				conn.Close()
				return
			}

			i := bytes.Index(data, []byte(" HTTP/"))
			if i < 0 {
				conn.Write([]byte("HTTP/1.1 400\r\n\r\n"))
				conn.Close()
				return
			}
			u, err := url.Parse(string(data[4:i]))
			if err != nil {
				conn.Write([]byte("HTTP/1.1 400\r\n\r\n"))
				conn.Close()
			}

			switch u.Path {
			case "/announce":
				w.jobQueue <- job{conn: conn, vals: u.Query()}
			case "/scrape":
				c := ctx{conn: conn, u: u}
				t.Scrape(&c)
				conn.Close()
			case "/":
				conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + string(index)))
				conn.Close()
			case "/dmca":
				conn.Write([]byte("HTTP/1.1 303\r\nLocation: https://www.youtube.com/watch?v=BwSts2s4ba4\r\n\r\n"))
				conn.Close()
			case "/stats":
				conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + shared.StatsHTML))
				conn.Close()
			default:
				conn.Write([]byte("HTTP/1.1 404\r\n\r\n"))
				conn.Close()
			}
		}()
	}
}

type job struct {
	conn net.Conn
	vals url.Values
}

type workers struct {
	tracker  *HTTPTracker
	jobQueue chan job
}

func (w *workers) consumeAnnounce() {
	for {
		select {
		case job := <-w.jobQueue:
			w.tracker.Announce(&job)
			job.conn.Close()
		}
	}
}
