package http

import (
	"net"
	"net/http"
)

type fakeRespWriter struct {
	conn net.Conn
}

func (w *fakeRespWriter) Header() http.Header {
	return make(http.Header)
}

func (w *fakeRespWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func (w *fakeRespWriter) WriteHeader(statusCode int) {}
