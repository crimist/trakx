package shared

type PeerID [20]byte
type PeerIP [4]byte

// Peer holds peer information stores in the database
type Peer struct {
	Complete bool
	IP       PeerIP
	Port     uint16
	LastSeen int64
}

func (db *PeerDatabase) makePeermap(h *Hash) (peermap *PeerMap) {
	// build struct and assign
	peermap = new(PeerMap)
	peermap.peers = make(map[PeerID]*Peer, 1)
	db.hashmap[*h] = peermap
	return
}

// Save writes a peer
func (db *PeerDatabase) Save(p *Peer, h *Hash, id *PeerID) {
	var dbPeer *Peer

	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		db.mu.Lock()
		peermap = db.makePeermap(h)
		db.mu.Unlock()
	}

	peermap.Lock()
	if !fast {
		dbPeer, ok = peermap.peers[*id]
	}
	peermap.peers[*id] = p
	peermap.Unlock()

	if !fast {
		if ok { // Already in db
			if dbPeer.Complete == false && p.Complete == true { // They completed
				AddExpval(&Expvar.Leeches, -1)
				AddExpval(&Expvar.Seeds, 1)
			}
			if dbPeer.Complete == true && p.Complete == false { // They uncompleted?
				AddExpval(&Expvar.Seeds, -1)
				AddExpval(&Expvar.Leeches, 1)
			}
			if dbPeer.IP != p.IP { // IP changed
				Expvar.IPs.Lock()
				Expvar.IPs.delete(dbPeer.IP)
				Expvar.IPs.inc(p.IP)
				Expvar.IPs.Unlock()
			}
		} else { // New
			Expvar.IPs.Lock()
			Expvar.IPs.inc(p.IP)
			Expvar.IPs.Unlock()
			if p.Complete {
				AddExpval(&Expvar.Seeds, 1)
			} else {
				AddExpval(&Expvar.Leeches, 1)
			}
		}
	}
}

// delete is like drop but doesn't lock
func (db *PeerDatabase) delete(p *Peer, h *Hash, id *PeerID) {
	peermap, ok := db.hashmap[*h]
	if !ok {
		return
	}
	delete(peermap.peers, *id)

	if !fast {
		if p.Complete {
			AddExpval(&Expvar.Seeds, -1)
		} else {
			AddExpval(&Expvar.Leeches, -1)
		}

		Expvar.IPs.Lock()
		Expvar.IPs.dec(p.IP)
		if Expvar.IPs.dead(p.IP) {
			Expvar.IPs.delete(p.IP)
		}
		Expvar.IPs.Unlock()
	}
}

// Drop deletes peer
func (db *PeerDatabase) Drop(p *Peer, h *Hash, id *PeerID) {
	db.mu.RLock()
	peermap, ok := db.hashmap[*h]
	db.mu.RUnlock()
	if !ok {
		return
	}

	peermap.Lock()
	delete(peermap.peers, *id)
	peermap.Unlock()

	if !fast {
		if p.Complete {
			AddExpval(&Expvar.Seeds, -1)
		} else {
			AddExpval(&Expvar.Leeches, -1)
		}

		Expvar.IPs.Lock()
		Expvar.IPs.dec(p.IP)
		if Expvar.IPs.dead(p.IP) {
			Expvar.IPs.delete(p.IP)
		}
		Expvar.IPs.Unlock()
	}
}
