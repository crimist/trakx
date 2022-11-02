package pools

import (
	"testing"
)

func TestPeerPool(t *testing.T) {
	Initialize(10)

	oldpeer := Peers.Get()
	oldpeer.Complete = true
	oldpeer.Port = 1234
	Peers.Put(oldpeer)
	newpeer := Peers.Get()

	if oldpeer != newpeer {
		t.Fatalf("failed to get same peer, oldpeer(%p) = %v newpeer(%p) = %v", oldpeer, oldpeer, newpeer, newpeer)
	}
}
