package http

import (
	"net"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/utils"
)

var (
	httpSuccessBytes = []byte("HTTP/1.1 200\r\n\r\n")
)

// string concats are optimized in go so these are faster than []byte appends etc.

func writeSuccess(c net.Conn, body string) {
	c.Write(utils.StringToBytesUnsafe("HTTP/1.1 200\r\n\r\n" + body))
}

func writeStatus(c net.Conn, status string) {
	c.Write(utils.StringToBytesUnsafe("HTTP/1.1 " + status + "\r\n\r\n"))
}

func writeFailure(conn net.Conn, msg string) {
	dictionary := pools.Dictionaries.Get()

	dictionary.String("failure reason", msg)
	conn.Write(append(httpSuccessBytes, dictionary.GetBytes()...))

	pools.Dictionaries.Put(dictionary)
}
