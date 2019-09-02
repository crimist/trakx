package inmemory

import "github.com/syc0x00/trakx/tracker/database"

func (db *Memory) InitExpvar() {
	if ok := db.Check(); !ok {
		panic("db not init before expvars")
	}

	// Called on main thread no locking needed
	for _, peermap := range db.hashmap {
		for _, peer := range peermap.peers {
			database.Expvar.IPs.M[peer.IP]++
			if peer.Complete == true {
				database.Expvar.Seeds++
			} else {
				database.Expvar.Leeches++
			}
		}
	}
}
