package shared

import (
	"go.uber.org/zap"
)

type PeerID [20]byte
type PeerIP [4]byte

// Peer holds peer information stores in the database
type Peer struct {
	IP       PeerIP
	Port     uint16
	Complete bool
	LastSeen int64
}

// Save creates or updates peer
func (p *Peer) Save(h Hash, id PeerID) {
	if Env == Dev {
		Logger.Info("Save",
			zap.Any("hash", h),
			zap.Any("peerid", id),
			zap.Any("Peer", p),
		)
	}

	// Create map if it doesn't exist
	if _, ok := PeerDB[h]; !ok {
		if Env == Dev {
			Logger.Info("Created hash map", zap.ByteString("hash", h[:]))
		}
		PeerDB[h] = make(map[PeerID]Peer)
	}

	key := expvarKey(h, id)
	peer, ok := PeerDB[h][id]
	if ok { // Exists
		if peer.Complete == false && p.Complete == true { // They completed
			delete(ExpvarLeeches, key)
			ExpvarSeeds[key] = true
		}
		if peer.Complete == true && p.Complete == false { // They uncompleted
			delete(ExpvarSeeds, key)
			ExpvarLeeches[key] = true
		}
		if peer.IP != p.IP { // IP changed
			delete(ExpvarIPs, peer.IP)
			ExpvarIPs[p.IP]++
		}
	} else { // Doesn't exist
		ExpvarIPs[p.IP]++
		if p.Complete {
			ExpvarSeeds[key] = true
		} else {
			ExpvarLeeches[key] = true
		}
	}

	PeerDB[h][id] = *p
}

// Delete deletes peer
func (p *Peer) Delete(h Hash, id PeerID) {
	if Env == Dev {
		Logger.Info("Delete",
			zap.Any("hash", h),
			zap.Any("peerid", id),
			zap.Any("Peer", p),
		)
	}

	key := expvarKey(h, id)
	delete(ExpvarSeeds, key)
	delete(ExpvarLeeches, key)
	ExpvarIPs[p.IP]--
	if num := ExpvarIPs[p.IP]; num < 1 {
		delete(ExpvarIPs, p.IP)
	}

	delete(PeerDB[h], id)
}
