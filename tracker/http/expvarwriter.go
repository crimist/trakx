package http

import (
	"net"
	"net/http"
)

var (
	expvarHeader = []byte("HTTP/1.1 200\r\nContent-Type: application/json; charset=utf-8\r\n\r\n")
)

type ExpvarResponseWriter struct {
	conn net.Conn
}

func (w ExpvarResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w ExpvarResponseWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func (w ExpvarResponseWriter) WriteHeader(statusCode int) {}
