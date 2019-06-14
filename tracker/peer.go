package tracker

import (
	"errors"
	"bytes"

	"go.uber.org/zap"
)

// Peer holds peer information stores in the database
type Peer struct {
	Key      []byte 
	Hash     Hash
	IP       string
	Port     uint16
	Complete bool  
	LastSeen int64 
}

// Save creates or updates peer
func (p *Peer) Save(id ID) error {
	logger.Info("Save",
		zap.Any("Peer", p),
	)

	db[id] = *p

	return nil
}

// Delete deletes peer
func (p *Peer) Delete(id ID) error {
	logger.Info("Delete",
		zap.Any("Peer", p),
	)
	
	if !bytes.Equal(db[id].Key, p.Key) {
		return errors.New("Invalid key")
	}
	
	delete(db, id)

	return nil
}
