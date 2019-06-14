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
	logger.Info("Save",
		zap.Any("hash", h),
		zap.Any("peerid", id),
		zap.Any("Peer", p),
	)

	// Create map if it doesn't exist
	if _, ok := db[h]; !ok {
		logger.Info("Created hash map", zap.ByteString("hash", h[:]))
		db[h] = make(map[PeerID]Peer)
	}

	db[h][id] = *p

	return nil
}

// Delete deletes peer
func (p *Peer) Delete(h Hash, id PeerID) error {
	logger.Info("Delete",
		zap.Any("Peer", p),
	)

	delete(db, h)
	return nil
}
