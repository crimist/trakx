package http

import (
	"net"
	"net/http"
)

var fakeHdr http.Header

type fakeRespWriter struct {
	conn net.Conn
}

func setupFakes() {
	fakeHdr = make(http.Header)
}

func (w *fakeRespWriter) Header() http.Header {
	return fakeHdr
}

func (w *fakeRespWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func (w *fakeRespWriter) WriteHeader(statusCode int) {}