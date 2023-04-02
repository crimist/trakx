package inmemory

import (
	"github.com/crimist/trakx/tracker/stats"
	"github.com/pkg/errors"
)

func (db *InMemory) SyncExpvars() error {
	if ok := db.Check(); !ok {
		return errors.New("driver not initiated before SyncExpvars")
	}

	var seeds, leeches int64

	// Called on main thread before thread/queue dispatch no locking needed
	for _, peermap := range db.hashes {
		for _, peer := range peermap.Peers {
			stats.IPStats.Inc(peer.IP)
			if peer.Complete {
				seeds++
			} else {
				leeches++
			}
		}
	}

	stats.Seeds.Store(seeds)
	stats.Leeches.Store(leeches)

	return nil
}
