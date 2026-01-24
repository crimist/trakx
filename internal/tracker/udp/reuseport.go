package udp

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// listenReusePort creates a UDP socket with SO_REUSEPORT enabled, allowing
// multiple sockets to bind to the same port for parallel packet processing.
func listenReusePort(network string, addr *net.UDPAddr) (*net.UDPConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var opErr error
			err := c.Control(func(fd uintptr) {
				opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})
			if err != nil {
				return err
			}
			return opErr
		},
	}
	conn, err := lc.ListenPacket(context.Background(), network, addr.String())
	if err != nil {
		return nil, err
	}
	return conn.(*net.UDPConn), nil
}
