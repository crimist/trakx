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
