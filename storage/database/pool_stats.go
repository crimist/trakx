package database

// Peerlist4PoolCreated returns the total number of pooled IPv4 peer lists created.
func (db *Database) Peerlist4PoolCreated() int64 {
	if db.peerLists == nil {
		return 0
	}
	created4, _ := db.peerLists.stats()
	return created4
}

// Peerlist6PoolCreated returns the total number of pooled IPv6 peer lists created.
func (db *Database) Peerlist6PoolCreated() int64 {
	if db.peerLists == nil {
		return 0
	}
	_, created6 := db.peerLists.stats()
	return created6
}
