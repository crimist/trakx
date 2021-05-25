package http

import (
	"fmt"
	"net"

	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/storage"
	"go.uber.org/zap"
)

const (
	httpRequestMax = 2600 // enough for scrapes up to 40 info_hashes
)

var (
	httpSuccess      = "HTTP/1.1 200\r\n\r\n"
	httpSuccessBytes = []byte(httpSuccess)
)

type HTTPTracker struct {
	peerdb   storage.Database
	workers  workers
	shutdown chan struct{}
}

// Init sets the HTTP trackers required values
func (t *HTTPTracker) Init(peerdb storage.Database) {
	t.peerdb = peerdb
	t.shutdown = make(chan struct{})
}

// Serve starts the HTTP service and begins to serve clients
func (t *HTTPTracker) Serve() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Conf.Tracker.HTTP.Port))
	if err != nil {
		config.Logger.Panic("net.Listen()", zap.Error(err))
	}

	t.workers = workers{
		tracker:  t,
		listener: ln,
	}

	t.workers.startWorkers(config.Conf.Tracker.HTTP.Threads)

	<-t.shutdown
	config.Logger.Info("Closing HTTP tracker listen socket")
	ln.Close()
}

// Shutdown gracefully closes the HTTP service by closing the listening connection
func (t *HTTPTracker) Shutdown() {
	if t == nil || t.shutdown == nil {
		return
	}
	var die struct{}
	t.shutdown <- die
}
