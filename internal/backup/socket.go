package backup

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	socketPrefix  = "trakx-backup-"
	acceptTimeout = 5 * time.Second
	streamTimeout = 2 * time.Minute
	dialTimeout   = 500 * time.Millisecond
)

// SocketSource requests a snapshot from a running daemon via SIGUSR1 + unix socket.
type SocketSource struct {
	ProcessID int
	CacheDir  string
}

// NewSocketSource creates a new socket source.
func NewSocketSource(processID int, cacheDir string) *SocketSource {
	return &SocketSource{
		ProcessID: processID,
		CacheDir:  cacheDir,
	}
}

func (s *SocketSource) Open() (io.ReadCloser, error) {
	socketPath, err := getSocketPath(s.CacheDir, s.ProcessID)
	if err != nil {
		return nil, err
	}

	zap.L().Debug("Preparing backup socket listener", zap.String("path", socketPath))
	if err := removeSocket(socketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to listen on backup socket")
	}

	if unixListener, ok := listener.(*net.UnixListener); ok {
		if err := unixListener.SetDeadline(time.Now().Add(acceptTimeout)); err != nil {
			listener.Close()
			removeSocket(socketPath)
			return nil, errors.Wrap(err, "failed to set accept timeout")
		}
	}

	zap.L().Debug("Signaling daemon for backup", zap.Int("pid", s.ProcessID))
	if err := syscall.Kill(s.ProcessID, syscall.SIGUSR1); err != nil {
		listener.Close()
		removeSocket(socketPath)
		return nil, errors.Wrap(err, "failed to signal daemon for backup")
	}

	zap.L().Debug("Waiting for daemon connection", zap.Duration("timeout", acceptTimeout))
	conn, err := listener.Accept()
	if err != nil {
		listener.Close()
		removeSocket(socketPath)
		return nil, errors.Wrap(err, "failed to accept daemon connection")
	}

	if unixConn, ok := conn.(*net.UnixConn); ok {
		if err := unixConn.SetDeadline(time.Now().Add(streamTimeout)); err != nil {
			conn.Close()
			listener.Close()
			removeSocket(socketPath)
			return nil, errors.Wrap(err, "failed to set stream timeout")
		}
	}

	zap.L().Debug("Streaming backup from daemon")

	// Return a wrapper that cleans up socket and listener on close
	return &socketReader{
		conn:       conn,
		listener:   listener,
		socketPath: socketPath,
	}, nil
}

func (s *SocketSource) Description() string {
	return fmt.Sprintf("socket:pid=%d", s.ProcessID)
}

type socketReader struct {
	conn       net.Conn
	listener   net.Listener
	socketPath string
}

func (r *socketReader) Read(p []byte) (n int, err error) {
	return r.conn.Read(p)
}

func (r *socketReader) Close() error {
	r.conn.Close()
	r.listener.Close()
	removeSocket(r.socketPath)
	return nil
}

// SocketDestination writes a snapshot to a unix socket (daemon side).
// This is used by the daemon when it receives SIGUSR1.
type SocketDestination struct {
	ProcessID int
	CacheDir  string
}

// NewSocketDestination creates a new socket destination.
func NewSocketDestination(processID int, cacheDir string) *SocketDestination {
	return &SocketDestination{
		ProcessID: processID,
		CacheDir:  cacheDir,
	}
}

func (d *SocketDestination) Create() (io.WriteCloser, error) {
	socketPath, err := getSocketPath(d.CacheDir, d.ProcessID)
	if err != nil {
		return nil, err
	}

	zap.L().Debug("Dialing backup socket", zap.String("path", socketPath))
	dialer := net.Dialer{Timeout: dialTimeout}
	conn, err := dialer.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}

	if unixConn, ok := conn.(*net.UnixConn); ok {
		if err := unixConn.SetDeadline(time.Now().Add(streamTimeout)); err != nil {
			conn.Close()
			return nil, err
		}
	}

	zap.L().Debug("Writing snapshot to socket")
	return conn, nil
}

func (d *SocketDestination) Description() string {
	return fmt.Sprintf("socket:pid=%d", d.ProcessID)
}

func getSocketPath(cacheDir string, processID int) (string, error) {
	return filepath.Join(cacheDir, fmt.Sprintf("%s%d.sock", socketPrefix, processID)), nil
}

func removeSocket(path string) error {
	zap.L().Debug("Removing socket path", zap.String("path", path))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove socket path")
	}
	return nil
}
