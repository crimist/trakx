package http

import (
	"net"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

func writeErr(conn net.Conn, msg string) {
	d := bencoding.GetDictionary()

	d.String("failure reason", msg)
	conn.Write(append(httpSuccessBytes, d.GetBytes()...))

	bencoding.PutDictionary(d)
}

func (t *HTTPTracker) clientError(conn net.Conn, msg string) {
	storage.Expvar.ClientErrors.Add(1)
	writeErr(conn, msg)
}

func (t *HTTPTracker) internalError(conn net.Conn, errmsg string, err error) {
	storage.Expvar.Errors.Add(1)
	writeErr(conn, "internal server error")
	config.Logger.Error(errmsg, zap.Error(err))
}
