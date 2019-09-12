package gomap

import "github.com/syc0x00/trakx/tracker/storage"

func (db *Memory) Expvar() {
	if ok := db.Check(); !ok {
		panic("db not init before expvars")
	}

	// Called on main thread no locking needed
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
