package http

import (
	"net"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/utils"
)

var (
	httpSuccess      = "HTTP/1.1 200 OK\r\n\r\n"
	httpSuccessBytes = []byte(httpSuccess)
)

// string concats are optimized in go so these are faster than []byte appends etc.

func writeSuccess(c net.Conn, body string) {
	c.Write(utils.StringToBytesUnsafe(httpSuccess + body))
}

func writeStatus(c net.Conn, status string) {
	c.Write(utils.StringToBytesUnsafe("HTTP/1.1 " + status + "\r\n\r\n"))
}

func (tracker *Tracker) error(conn net.Conn, msg string) {
	tracker.collector.ErrorResponse()

	dictionary := bencoding.AcquireDictionary()

	dictionary.String("failure reason", msg)
	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))

	bencoding.ReleaseDictionary(dictionary)
}
