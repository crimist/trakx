package http

import (
	"fmt"
	"net"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/pkg/errors"
)

const (
	httpRequestMax = 2600 // enough for scrapes up to 40 info_hashes
)

type HTTPTracker struct {
	peerdb   storage.Database
	workers  workers
	shutdown chan struct{}
}

// Init sets up the HTTPTracker.
func (t *HTTPTracker) Init(peerdb storage.Database) {
	t.peerdb = peerdb
	t.shutdown = make(chan struct{})
}

// Serve begins listening and serving clients.
func (t *HTTPTracker) Serve() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%v:%v", config.Config.HTTP.IP, config.Config.HTTP.Port))
	if err != nil {
		return errors.Wrap(err, "Failed to open TCP listen socket")
	}

	cache, err := config.GenerateEmbeddedCache()
	if err != nil {
		return errors.Wrap(err, "failed to generate embedded cache")
	}

	t.workers = workers{
		tracker:   t,
		listener:  ln,
		fileCache: cache,
	}

	t.workers.startWorkers(config.Config.HTTP.Threads)

	<-t.shutdown
	if err := ln.Close(); err != nil {
		return errors.Wrap(err, "Failed to close tcp listen socket")
	}

	return nil
}

// Shutdown stops the HTTP tracker server by closing the socket.
func (t *HTTPTracker) Shutdown() {
	if t == nil || t.shutdown == nil {
		return
	}
	var die struct{}
	t.shutdown <- die
}
