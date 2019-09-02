package http

import (
	"net"

	"github.com/syc0x00/trakx/bencoding"
	"github.com/syc0x00/trakx/tracker/database"
	"go.uber.org/zap"
)

func writeErr(conn net.Conn, msg string) {
	d := bencoding.NewDict()
	d.String("failure reason", msg)
	conn.Write([]byte("HTTP/1.1 200\r\n\r\n" + d.Get()))
}

func (t *HTTPTracker) clientError(conn net.Conn, msg string) {
	database.AddExpval(&database.Expvar.Clienterrs, 1)
	writeErr(conn, msg)
}

func (t *HTTPTracker) internalError(conn net.Conn, errmsg string, err error) {
	database.AddExpval(&database.Expvar.Errs, 1)
	writeErr(conn, "internal server error")
	t.logger.Error(errmsg, zap.Error(err))
}
