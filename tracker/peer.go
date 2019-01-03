package tracker

import (
	"bytes"
	"errors"

	"go.uber.org/zap"
)

// Peer :clap:
type Peer struct {
	ID       []byte `gorm:"primary_key;unique;not_null"`
	Key      []byte
	Hash     []byte
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

func (p *Peer) checkKey() error {
	// Check if key
	pDB := Peer{}
	db.Where("id = ?", p.ID).First(&pDB)
	if p.Key != nil {
		if bytes.Equal(p.Key, pDB.Key) == false {
			logger.Info("invalid key",
				zap.String("ip", p.IP),
				zap.ByteString("real key", pDB.Key),
				zap.ByteString("provided key", p.Key),
			)
			return errors.New("Invalid key")
		}
	}
	return nil
}

// Save creates or updates peer
func (p *Peer) Save() error {
	logger.Info("Save",
		zap.Any("Peer", p),
	)

	// Create it if not exist
	if err := db.FirstOrCreate(p).Error; err != nil {
		return err
	}
	if err := p.checkKey(); err != nil {
		return err
	}
	return db.Save(p).Error
}

// Delete deletes peer
func (p *Peer) Delete() error {
	logger.Info("Delete",
		zap.Any("Peer", p),
	)
	if err := p.checkKey(); err != nil {
		return err
	}
	return db.Delete(p).Error
}
