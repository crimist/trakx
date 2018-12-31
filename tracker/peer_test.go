package tracker_test

import (
	"testing"
	"time"

	"github.com/Syc0x00/Trakx/tracker"
	"github.com/jinzhu/gorm"
)

func TestSavePeer(t *testing.T) {
	db, err := gorm.Open("mysql", "root@/bittorrent")
	if err != nil {
		t.Error(err)
	}
	db.LogMode(true) // Enable extensive logging

	p := tracker.Peer{
		ID:       "id",
		Key:      "key",
		Hash:     "hash",
		IP:       "127.0.0.1",
		Port:     1234,
		Complete: false,
		LastSeen: time.Now().Unix(),
	}

	db.Create(&p)

}
