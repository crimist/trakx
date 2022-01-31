package http

import (
	"fmt"
	"net"

	"github.com/crimist/trakx/tracker/config"
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
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Conf.Tracker.HTTP.Port))
	if err != nil {
		return errors.Wrap(err, "Failed to open TCP listen socket")
	}

	t.workers = workers{
		tracker:  t,
		listener: ln,
	}

	t.workers.startWorkers(config.Conf.Tracker.HTTP.Threads)

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
