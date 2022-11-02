package http

import (
	"net"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/stats"
	"go.uber.org/zap"
)

func writeErr(conn net.Conn, msg string) {
	dictionary := pools.Dictionaries.Get()

	dictionary.String("failure reason", msg)
	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))

	pools.Dictionaries.Put(dictionary)
}

func (t *HTTPTracker) clientError(conn net.Conn, msg string) {
	stats.ClientErrors.Add(1)
	writeErr(conn, msg)
}

func (t *HTTPTracker) internalError(conn net.Conn, errmsg string, err error) {
	stats.ServerErrors.Add(1)
	writeErr(conn, "internal server error")
	config.Logger.Error(errmsg, zap.Error(err))
}
