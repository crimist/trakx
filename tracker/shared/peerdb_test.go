package shared

import (
	"testing"
)

func TestCheck(t *testing.T) {
	var db PeerDatabase
	if db.check() != false {
		t.Error("check() on empty db returned true")
	}
}

func TestTrim(t *testing.T) {
	var c Config
	c.Database.Peer.Timeout = 0

	db := dbWithHashes(1000000)
	db.conf = &c

	db.trim()

	db, _ = dbWithPeers(1000000)
	db.conf = &c

	db.trim()
}
