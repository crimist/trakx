package http

import (
	"net"
	"net/http"
)

var (
	fakeHdr     http.Header
	statsHeader = []byte("HTTP/1.1 200\r\nContent-Type: application/json; charset=utf-8\r\n\r\n")
)

func init() {
	fakeHdr = make(http.Header)
}

type fakeRespWriter struct {
	conn net.Conn
}

func (w fakeRespWriter) Header() http.Header {
	return fakeHdr
}

func (w fakeRespWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func (w fakeRespWriter) WriteHeader(statusCode int) {}
