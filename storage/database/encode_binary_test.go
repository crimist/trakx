package database

import (
	"fmt"
	"math/rand"
	"net/netip"
	"reflect"
	"testing"
	"time"

	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
)

func TestBinaryCoder(t *testing.T) {
	db, err := NewDatabase(Config{
		InitalSize:        1,
		EvictionFrequency: 1 * time.Minute,
		ExpirationTime:    1 * time.Minute,
		Collector:         stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}
	testPeer := storage.Peer{
		Complete: true,
		IP:       testPeerIP,
		Port:     1234,
	}
	db.PeerAdd(testTorrentHash1, testPeerID1, testPeer.IP, testPeer.Port, testPeer.Complete)

	data, err := encodeBinary(db)
	if err != nil {
		t.Fatal("encodeBinary threw error: ", err)
	}
	oldtorrents := db.torrents
	db, err = NewDatabase(Config{
		InitalSize:        1,
		EvictionFrequency: 1 * time.Minute,
		ExpirationTime:    1 * time.Minute,
		Collector:         stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}
	if _, _, err = decodeBinary(db, data); err != nil {
		t.Fatal("decodeBinary threw error: ", err)
	}

	if _, ok := db.torrents[testTorrentHash1]; !ok {
		t.Fatal("torrent missing peer")
	}
	if db.torrents[testTorrentHash1].Seeds != oldtorrents[testTorrentHash1].Seeds {
		t.Fatalf("seeds = %v, want %v", db.torrents[testTorrentHash1].Seeds, oldtorrents[testTorrentHash1].Seeds)
	}
	if db.torrents[testTorrentHash1].Leeches != oldtorrents[testTorrentHash1].Leeches {
		t.Fatalf("leeches = %v, want %v", db.torrents[testTorrentHash1].Leeches, oldtorrents[testTorrentHash1].Leeches)
	}
	if !reflect.DeepEqual(db.torrents[testTorrentHash1].Peers, oldtorrents[testTorrentHash1].Peers) {
		t.Fatalf("peers = %v, want %v", db.torrents[testTorrentHash1].Peers, oldtorrents[testTorrentHash1].Peers)
	}
}

// Helper function to create a database with realistic peer distribution
func createTestDatabase(totalPeers int, rnd *rand.Rand) *Database {
	db, err := NewDatabase(Config{
		InitalSize:        1,
		EvictionFrequency: 1 * time.Hour,
		ExpirationTime:    1 * time.Hour,
		Collector:         stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		panic("Failed to create database: " + err.Error())
	}

	// Calculate number of torrents based on peer count
	// Average ~100 peers per torrent, but will vary with power-law distribution
	numTorrents := max(1, totalPeers/100)
	if numTorrents > totalPeers {
		numTorrents = totalPeers
	}

	// Generate torrent hashes
	torrents := make([]storage.Hash, numTorrents)
	for i := range torrents {
		rnd.Read(torrents[i][:])
	}

	// Use Zipf distribution for peer allocation (power-law distribution)
	zipf := rand.NewZipf(rnd, 1.1, 1.0, uint64(numTorrents-1))

	peersAdded := 0
	maxPeersPerTorrent := 5000

	for peersAdded < totalPeers {
		// Select a torrent using Zipf distribution
		torrentIdx := int(zipf.Uint64())
		if torrentIdx >= len(torrents) {
			torrentIdx = len(torrents) - 1
		}

		hash := torrents[torrentIdx]

		// Get current peer count for this torrent
		torrent, exists := db.torrents[hash]
		currentPeerCount := 0
		if exists {
			currentPeerCount = len(torrent.Peers)
		}

		// Skip if torrent already at max capacity
		if currentPeerCount >= maxPeersPerTorrent {
			continue
		}

		// Add a peer
		var peerid storage.PeerID
		rnd.Read(peerid[:])

		// Realistic seed/leech ratio: ~20% seeds, 80% leeches
		isComplete := rnd.Float32() < 0.20

		db.PeerAdd(hash, peerid, testPeerIP, uint16(1024+rnd.Intn(64512)), isComplete)
		peersAdded++
	}

	return db
}

// Test encoding/decoding with various peer counts
func TestBinaryCoderScaling(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	peerCounts := []int{10, 100, 1000, 10000}

	for _, peers := range peerCounts {
		t.Run(fmt.Sprintf("%d_peers", peers), func(t *testing.T) {
			originalDB := createTestDatabase(peers, rnd)

			// Count original peers
			originalPeerCount := 0
			originalTorrentCount := 0
			for _, torrent := range originalDB.torrents {
				originalPeerCount += len(torrent.Peers)
				originalTorrentCount++
			}

			// Encode
			data, err := encodeBinary(originalDB)
			if err != nil {
				t.Fatal(err)
			}

			// Decode
			newDB, _ := NewDatabase(Config{
				InitalSize:        1,
				EvictionFrequency: 1 * time.Hour,
				ExpirationTime:    1 * time.Hour,
				Collector:         stats.NewCollectors(false, false, 0),
			})

			numPeers, numTorrents, err := decodeBinary(newDB, data)
			if err != nil {
				t.Fatal(err)
			}

			// Verify counts
			if numPeers != originalPeerCount {
				t.Errorf("Peer count mismatch: expected %d, got %d", originalPeerCount, numPeers)
			}
			if numTorrents != originalTorrentCount {
				t.Errorf("Torrent count mismatch: expected %d, got %d", originalTorrentCount, numTorrents)
			}

			// Verify data integrity
			for hash, originalTorrent := range originalDB.torrents {
				newTorrent, exists := newDB.torrents[hash]
				if !exists {
					t.Errorf("Missing torrent %v", hash)
					continue
				}

				if newTorrent.Seeds != originalTorrent.Seeds {
					t.Errorf("Seeds mismatch for torrent %v: expected %d, got %d",
						hash, originalTorrent.Seeds, newTorrent.Seeds)
				}

				if len(newTorrent.Peers) != len(originalTorrent.Peers) {
					t.Errorf("Peer count mismatch for torrent %v: expected %d, got %d",
						hash, len(originalTorrent.Peers), len(newTorrent.Peers))
				}
			}

			t.Logf("Encoded %d peers across %d torrents into %d bytes (%.1f bytes/peer)",
				numPeers, numTorrents, len(data), float64(len(data))/float64(numPeers))
		})
	}
}

// Benchmark encoding at various scales
func BenchmarkEncode(b *testing.B) {
	rnd := rand.New(rand.NewSource(42))

	for peers := 1000; peers <= 5000000; peers *= 10 {
		b.Run(fmt.Sprintf("%d_peers", peers), func(b *testing.B) {
			db := createTestDatabase(peers, rnd)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := encodeBinary(db)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark decoding at various scales
func BenchmarkDecode(b *testing.B) {
	rnd := rand.New(rand.NewSource(42))

	for peers := 1000; peers <= 5000000; peers *= 10 {
		b.Run(fmt.Sprintf("%d_peers", peers), func(b *testing.B) {
			db := createTestDatabase(peers, rnd)
			data, err := encodeBinary(db)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				newDB, _ := NewDatabase(Config{
					InitalSize:        1,
					EvictionFrequency: 1 * time.Hour,
					ExpirationTime:    1 * time.Hour,
					Collector:         stats.NewCollectors(false, false, 0),
				})
				_, _, err := decodeBinary(newDB, data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark full roundtrip
func BenchmarkRoundtrip(b *testing.B) {
	rnd := rand.New(rand.NewSource(42))

	for peers := 1000; peers <= 1000000; peers *= 10 {
		b.Run(fmt.Sprintf("%d_peers", peers), func(b *testing.B) {
			db := createTestDatabase(peers, rnd)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				data, err := encodeBinary(db)
				if err != nil {
					b.Fatal(err)
				}

				newDB, _ := NewDatabase(Config{
					InitalSize:        1,
					EvictionFrequency: 1 * time.Hour,
					ExpirationTime:    1 * time.Hour,
					Collector:         stats.NewCollectors(false, false, 0),
				})
				_, _, err = decodeBinary(newDB, data)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Test size efficiency
func TestSizeEfficiency(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	peerCounts := []int{1000, 10000, 100000, 1000000}

	t.Log("Binary Encoding Size Efficiency:")
	t.Log("=================================")
	t.Logf("%-12s | %-15s | %-15s | %-15s", "Peer Count", "Size (KB)", "Bytes/Peer", "Torrents")
	t.Log("-------------|-----------------|-----------------|-----------------|")

	for _, peers := range peerCounts {
		db := createTestDatabase(peers, rnd)

		data, err := encodeBinary(db)
		if err != nil {
			t.Fatal(err)
		}

		sizeKB := float64(len(data)) / 1024
		bytesPerPeer := float64(len(data)) / float64(peers)
		numTorrents := len(db.torrents)

		t.Logf("%-12d | %15.1f | %15.1f | %15d",
			peers, sizeKB, bytesPerPeer, numTorrents)
	}
}

// Test database distribution to verify realistic load
func TestDatabaseDistribution(t *testing.T) {
	rnd := rand.New(rand.NewSource(42))
	peers := 100000

	db := createTestDatabase(peers, rnd)

	// Collect peer counts per torrent
	peerCounts := make([]int, 0)
	totalPeers := 0
	for _, torrent := range db.torrents {
		count := len(torrent.Peers)
		peerCounts = append(peerCounts, count)
		totalPeers += count
	}

	// Calculate statistics
	var sum, max, min int
	min = peerCounts[0]
	for _, count := range peerCounts {
		sum += count
		if count > max {
			max = count
		}
		if count < min {
			min = count
		}
	}
	avg := float64(sum) / float64(len(peerCounts))

	t.Logf("Database Distribution Statistics:")
	t.Logf("  Total Torrents: %d", len(db.torrents))
	t.Logf("  Total Peers: %d", totalPeers)
	t.Logf("  Avg Peers/Torrent: %.1f", avg)
	t.Logf("  Max Peers/Torrent: %d", max)
	t.Logf("  Min Peers/Torrent: %d", min)

	// Show distribution buckets
	buckets := map[string]int{
		"1-10":      0,
		"11-50":     0,
		"51-100":    0,
		"101-500":   0,
		"501-1000":  0,
		"1001-5000": 0,
	}

	for _, count := range peerCounts {
		switch {
		case count <= 10:
			buckets["1-10"]++
		case count <= 50:
			buckets["11-50"]++
		case count <= 100:
			buckets["51-100"]++
		case count <= 500:
			buckets["101-500"]++
		case count <= 1000:
			buckets["501-1000"]++
		default:
			buckets["1001-5000"]++
		}
	}

	t.Logf("\n  Peer Distribution:")
	t.Logf("    1-10 peers:      %d torrents", buckets["1-10"])
	t.Logf("    11-50 peers:     %d torrents", buckets["11-50"])
	t.Logf("    51-100 peers:    %d torrents", buckets["51-100"])
	t.Logf("    101-500 peers:   %d torrents", buckets["101-500"])
	t.Logf("    501-1000 peers:  %d torrents", buckets["501-1000"])
	t.Logf("    1001-5000 peers: %d torrents", buckets["1001-5000"])

	// Verify we have a power-law distribution (most torrents have few peers)
	if buckets["11-50"] < buckets["101-500"] {
		t.Error("Distribution doesn't follow expected power-law pattern")
	}
}

// Test IPv4 and IPv6 encoding
func TestIPv4AndIPv6(t *testing.T) {
	db, err := NewDatabase(Config{
		InitalSize:        1,
		EvictionFrequency: 1 * time.Minute,
		ExpirationTime:    1 * time.Minute,
		Collector:         stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		t.Fatal("Failed to create database")
	}

	// Add IPv4 peer
	ipv4 := testPeerIP // Already IPv4
	db.PeerAdd(testTorrentHash1, testPeerID1, ipv4, 1234, true)

	// Add IPv6 peer
	ipv6 := mustParseIP("2001:db8::1")
	db.PeerAdd(testTorrentHash1, testPeerID2, ipv6, 5678, false)

	// Encode
	data, err := encodeBinary(db)
	if err != nil {
		t.Fatal(err)
	}

	// Decode
	newDB, _ := NewDatabase(Config{
		InitalSize:        1,
		EvictionFrequency: 1 * time.Minute,
		ExpirationTime:    1 * time.Minute,
		Collector:         stats.NewCollectors(false, false, 0),
	})
	numPeers, _, err := decodeBinary(newDB, data)
	if err != nil {
		t.Fatal(err)
	}

	if numPeers != 2 {
		t.Fatalf("Expected 2 peers, got %d", numPeers)
	}

	// Verify IPv4 peer
	peer1 := newDB.torrents[testTorrentHash1].Peers[testPeerID1]
	if peer1.IP.Compare(ipv4) != 0 {
		t.Errorf("IPv4 mismatch: expected %v, got %v", ipv4, peer1.IP)
	}
	if peer1.Port != 1234 {
		t.Errorf("IPv4 port mismatch: expected 1234, got %d", peer1.Port)
	}

	// Verify IPv6 peer
	peer2 := newDB.torrents[testTorrentHash1].Peers[testPeerID2]
	if peer2.IP.Compare(ipv6) != 0 {
		t.Errorf("IPv6 mismatch: expected %v, got %v", ipv6, peer2.IP)
	}
	if peer2.Port != 5678 {
		t.Errorf("IPv6 port mismatch: expected 5678, got %d", peer2.Port)
	}

	// Check size efficiency: IPv4 should be smaller
	t.Logf("Total encoded size for 1 IPv4 + 1 IPv6 peer: %d bytes", len(data))
}

// Benchmark memory allocations
func BenchmarkAllocations(b *testing.B) {
	rnd := rand.New(rand.NewSource(42))
	db := createTestDatabase(10000, rnd)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := encodeBinary(db)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper to parse IP for testing
func mustParseIP(s string) netip.Addr {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return addr
}
