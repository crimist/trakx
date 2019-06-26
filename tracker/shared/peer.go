package shared

import (
	"go.uber.org/zap"
)

type PeerID [20]byte

type UDPPeer struct {
	IP   int32
	Port uint16
}

// Peer holds peer information stores in the database
type Peer struct {
	Key      []byte
	IP       string
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

	// !x
	peer, ok := PeerDB[h][id]
	if ok { // Exists
		if peer.Complete == false && p.Complete == true { // They completed
			delete(ExpvarLeeches, p.IP)
			ExpvarSeeds[p.IP] = true
		}
		if peer.Complete == true && p.Complete == false { // They uncompleted
			delete(ExpvarSeeds, p.IP)
			ExpvarLeeches[p.IP] = true
		}
		if peer.IP != p.IP { // IP changed
			delete(ExpvarIPs, peer.IP)
			ExpvarIPs[p.IP] = true
		}
	} else { // Doesn't exist
		ExpvarIPs[p.IP] = true
		if p.Complete {
			ExpvarSeeds[p.IP] = true
		} else {
			ExpvarLeeches[p.IP] = true
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

	// !x
	delete(ExpvarIPs, p.IP)
	delete(ExpvarSeeds, p.IP)
	delete(ExpvarLeeches, p.IP)

	delete(PeerDB[h], id)
}
