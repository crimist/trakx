package tracker

import (
	"go.uber.org/zap"
)

type PeerID [20]byte

// Peer holds peer information stores in the database
type Peer struct {
	Key      []byte
	IP       string
	Port     uint16
	Complete bool
	LastSeen int64
}

// Save creates or updates peer
func (p *Peer) Save(h Hash, id PeerID) error {
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

	PeerDB[h][id] = *p

	return nil
}

// Delete deletes peer
func (p *Peer) Delete(h Hash, id PeerID) error {
	if Env == Dev {
		Logger.Info("Delete",
			zap.Any("hash", h),
			zap.Any("peerid", id),
			zap.Any("Peer", p),
		)
	}

	delete(PeerDB[h], id)
	return nil
}
