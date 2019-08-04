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

	for {
		conn, err := ln.Accept()
		if err != nil {
			t.logger.Info("net.Listen()", zap.Error(err))
			continue
		}
		go func() {
			b := pool.Get().([]byte)
			defer func() {
				conn.Close()
				for i := range b {
					b[i] = 0
				}
				pool.Put(b)
			}()

			conn.Read(b)

			// Do stuff
			if !bytes.Equal(b[0:4], []byte("GET ")) {
				conn.Write([]byte("Trakx only supports GET"))
				return
			}

			i := bytes.Index(b, []byte(" HTTP/"))
			if i < 0 {
				conn.Write([]byte("HTTP/1.1 400\r\n\r\n"))
				return
			}
			u, _ := url.Parse(string(b[4:i]))

			c := ctx{conn: conn, u: u}

			switch u.Path {
			case "/":
				c.WriteHTTP("200", string(index))
			case "/announce":
				t.Announce(&c)
			case "/scrape":
				t.Scrape(&c)
			case "/dmca":
				conn.Write([]byte("HTTP/1.1 303\r\nLocation: https://www.youtube.com/watch?v=BwSts2s4ba4\r\n\r\n"))
			case "/stats":
				c.WriteHTTP("200", string(shared.StatsHTML))
			default:
				c.WriteHTTP("404", "")
			}
		}()
	}
}

func (c *ctx) WriteHTTP(status string, msg string) {
	c.conn.Write([]byte("HTTP/1.1 " + status + "\r\n\r\n" + msg))
}
