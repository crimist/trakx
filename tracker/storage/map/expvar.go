package gomap

import "github.com/crimist/trakx/tracker/storage"

func (db *Memory) Expvar() {
	if ok := db.Check(); !ok {
		db.logger.Fatal("storage was not initialized before calling Expvar()")
		return
	}

	// Called on main thread before thread dispatch no locking needed
	for _, peermap := range db.hashmap {
		for _, peer := range peermap.peers {
			storage.Expvar.IPs.M[peer.IP]++
			if peer.Complete == true {
				storage.Expvar.Seeds++
			} else {
				storage.Expvar.Leeches++
			}
		}
	}
}
