package tracker

import (
	"go.uber.org/zap"
)

// Peer :clap:
type Peer struct {
	ID       string
	PeerKey  string `gorm:"primary_key;unique"`
	Hash     string
	IP       string
	Port     uint16
	Complete bool
	LastSeen int64
}

// InitPeer creates the peer db if it doesn't exist
func initPeer() {
	if db.HasTable(&Peer{}) == false {
		db.CreateTable(&Peer{})
	}
}

// Save creates or updates peer
func (p *Peer) Save() error {
	logger.Info("Save",
		zap.Any("Peer", p),
	)
	return db.FirstOrCreate(p).Save(p).Error
}

// Delete deletes peer
func (p *Peer) Delete() error {
	logger.Info("Delete",
		zap.Any("Peer", p),
	)
	return db.Delete(&p).Error
}
