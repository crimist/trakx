package http

import (
	"net"
	"sync/atomic"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

func writeErr(conn net.Conn, msg string) {
	d := bencoding.NewDict()
	d.Add("failure reason", msg)
	conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + d.Get()))
}

func (t *HTTPTracker) clientError(conn net.Conn, msg string) {
	atomic.AddInt64(&shared.Expvar.Clienterrs, 1)
	writeErr(conn, msg)
}

func (t *HTTPTracker) internalError(conn net.Conn, errmsg string, err error) {
	atomic.AddInt64(&shared.Expvar.Errs, 1)
	writeErr(conn, "internal server error")
	t.logger.Error(errmsg, zap.Error(err))
}
