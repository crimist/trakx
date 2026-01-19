package database

import (
	"net/netip"

	"github.com/crimist/trakx/internal/storage"
)

var (
	testTorrentHash1 = storage.Hash([20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	testTorrentHash2 = storage.Hash([20]byte{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	testPeerIP       = netip.AddrFrom4([4]byte{1, 2, 3, 4})
	testPeerID1      = storage.PeerID([20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	testPeerID2      = storage.PeerID([20]byte{1, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
)
