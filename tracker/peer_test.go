package tracker_test

import (
	"testing"
	"time"

	"github.com/Syc0x00/Trakx/tracker"
	"github.com/jinzhu/gorm"
)

func TestSavePeer(t *testing.T) {
	t.Skip()
	db, err := gorm.Open("mysql", "root@/bittorrent")
	if err != nil {
		t.Error(err)
	}
	db.LogMode(true) // Enable extensive logging

	p0 := tracker.Peer{
		ID:       []byte{'I', 'D', 0x00},
		Hash:     []byte{'H', 'A', 'S', 'H'},
		IP:       "127.0.0.1",
		Port:     1234,
		Complete: false,
		LastSeen: time.Now().Unix(),
	}

	db.Create(&p0)

	p1 := tracker.Peer{
		ID:       []byte{'I', 'D', 0x00},
		Hash:     []byte{'H', 'A', 'S', 'H'},
		IP:       "127.0.0.1",
		Port:     1234,
		Complete: false,
		LastSeen: time.Now().Unix(),
	}

	if err := db.Create(&p1).Error; err == nil {
		t.Fatal("Duplicate ID allowed", err)
	}

	p2 := tracker.Peer{
		ID:       []byte{0x01, 0x02, 0x03, 0x04},
		Key:      []byte{0x0A, 0x0B, 0x0C, 0x0D},
		Hash:     []byte{'H', 'A', 'S', 'H'},
		IP:       "127.0.0.1",
		Port:     1234,
		Complete: false,
		LastSeen: time.Now().Unix(),
	}

	if err := db.Create(&p2).Error; err != nil {
		t.Error(err)
	}

	db.Delete(&p0)
	db.Delete(&p1)
	db.Delete(&p2)
}
