package inmemory

func (db *InMemory) syncExpvars() {
	var seeds, leeches int64

	// Called on main thread before thread/queue dispatch no locking needed
	for _, peermap := range db.torrents {
		for _, peer := range peermap.Peers {
			db.stats.IPStats.Inc(peer.IP)
			if peer.Complete {
				seeds++
			} else {
				leeches++
			}
		}
	}

	db.stats.Seeds.Store(seeds)
	db.stats.Leeches.Store(leeches)
}
