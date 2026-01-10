package database

import (
	"fmt"
	"net/netip"
	"testing"

	"github.com/crimist/trakx/storage"
)

const benchHashPeerCount = 1 << 12

var (
	sinkPeers  [][]byte
	sinkV4Len  int
	sinkV6Len  int
	sinkSeeds  uint16
	sinkLeeches uint16
)

func makeBenchAddrs(n int) []netip.Addr {
	addrs := make([]netip.Addr, n)
	for i := range addrs {
		if i&1 == 0 {
			addrs[i] = netip.AddrFrom4([4]byte{10, 0, 0, byte(i%250 + 1)})
			continue
		}
		var ip [16]byte
		ip[0] = 0x20
		ip[1] = 0x01
		ip[2] = 0x0d
		ip[3] = 0xb8
		ip[15] = byte(i)
		addrs[i] = netip.AddrFrom16(ip)
	}
	return addrs
}

func prefillTorrent(db *Database, hash storage.Hash, ids []storage.PeerID, addrs []netip.Addr) {
	for i, id := range ids {
		ip := addrs[i]
		complete := (i & 1) == 0
		db.PeerAdd(hash, id, ip, 1234, complete)
	}
}

func BenchmarkTorrentPeers(b *testing.B) {
	ids := makePeerIDs(benchHashPeerCount, 21)
	addrs := makeBenchAddrs(len(ids))
	numWants := []uint{20, 50, 100, 200}
	modes := []struct {
		label   string
		include bool
	}{
		{label: "NoPeerID", include: false},
		{label: "WithPeerID", include: true},
	}

	for _, mode := range modes {
		for _, numWant := range numWants {
			numWant := numWant
			b.Run(fmt.Sprintf("%s/NumWant_%d", mode.label, numWant), func(b *testing.B) {
				db := newBenchDB(b)
				prefillTorrent(db, testTorrentHash1, ids, addrs)

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					sinkPeers = db.TorrentPeers(testTorrentHash1, numWant, mode.include)
				}
			})
		}
	}
}

func BenchmarkTorrentPeersCompact(b *testing.B) {
	ids := makePeerIDs(benchHashPeerCount, 22)
	addrs := makeBenchAddrs(len(ids))
	numWants := []uint{20, 50, 100, 200}
	modes := []struct {
		label    string
		wantedIP storage.IPVersion
	}{
		{label: "IPv4", wantedIP: storage.IPv4},
		{label: "IPv6", wantedIP: storage.IPv6},
		{label: "IPv4IPv6", wantedIP: storage.IPv4 | storage.IPv6},
	}

	for _, mode := range modes {
		for _, numWant := range numWants {
			numWant := numWant
			b.Run(fmt.Sprintf("%s/NumWant_%d", mode.label, numWant), func(b *testing.B) {
				db := newBenchDB(b)
				prefillTorrent(db, testTorrentHash1, ids, addrs)

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					peers := db.TorrentPeersCompact(testTorrentHash1, numWant, mode.wantedIP)
					sinkV4Len = len(peers.V4)
					sinkV6Len = len(peers.V6)
					peers.Release()
				}
			})
		}
	}
}

func BenchmarkTorrentStats(b *testing.B) {
	ids := makePeerIDs(benchHashPeerCount, 23)
	addrs := makeBenchAddrs(len(ids))
	db := newBenchDB(b)
	prefillTorrent(db, testTorrentHash1, ids, addrs)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkSeeds, sinkLeeches = db.TorrentStats(testTorrentHash1)
	}
}
