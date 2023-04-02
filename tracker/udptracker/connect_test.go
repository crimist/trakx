package udptracker

import (
	"testing"
	"time"

	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/storage/inmemory"
	"github.com/crimist/trakx/tracker/udptracker/conncache"
)

// TODO: resume here, writing tests for UDP tracker
// need to refactor the peer database stuff to be able to use it

func TestConnect(t *testing.T) {
	peerDB, err := inmemory.NewInMemory(1, nil, storage.Config{})
	if err != nil {
		t.Error("failed to create database", err)
	}
	connCache := conncache.NewConnectionCache(1, 1*time.Minute, 1*time.Minute, "")
	NewTracker(peerDB, connCache, TrackerConfig{})
}
