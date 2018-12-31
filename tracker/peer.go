package tracker

import (
	"go.uber.org/zap"
)

// Peer :clap:
type Peer struct {
	ID       string `gorm:"primary_key;unique"`
	Key      string
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
	/* // Better?
	if err := db.FirstOrCreate(p).Error; err != nil {
		return err
	}
	return db.Where("id = ? AND peer_key = ?", p.ID, p.PeerKey).Save(p).Error
	*/
}

// Delete deletes peer
func (p *Peer) Delete() error {
	logger.Info("Delete",
		zap.Any("Peer", p),
	)
	return db.Delete(&p).Error
}
