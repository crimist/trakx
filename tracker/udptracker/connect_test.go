package udptracker

import (
	"testing"
	"time"

	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/tracker/udptracker/conncache"
)

// TODO: resume here, writing tests for UDP tracker
// need to refactor the peer database stuff to be able to use it

func TestConnect(t *testing.T) {
	peerDB, err := storage.Open()
	if err != nil {
		t.Error("Failed to initialize storage", err)
	}
	connCache := conncache.NewConnectionCache(1, 1*time.Minute, 1*time.Minute, "")
	NewTracker(peerDB, connCache, TrackerConfig{})
}
