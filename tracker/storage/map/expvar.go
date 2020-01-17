package gomap

import (
	"github.com/crimist/trakx/tracker/storage"
	"github.com/pkg/errors"
)

func (db *Memory) Expvar() error {
	if ok := db.Check(); !ok {
		return errors.New("driver not init before calling Expvar()")
	}

	// Called on main thread before thread/queue dispatch no locking needed
	for _, peermap := range db.hashmap {
		for _, peer := range peermap.peers {
			storage.Expvar.IPs.Inc(peer.IP)
			if peer.Complete == true {
				storage.Expvar.Seeds++
			} else {
				storage.Expvar.Leeches++
			}
		}
	}

	return nil
}
