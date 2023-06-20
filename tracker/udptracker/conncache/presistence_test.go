package conncache

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/netip"
	"os"
	"testing"
	"time"
)

func TestPersistence(t *testing.T) {
	const persistancePath = "cache_entries.tmp"
	const id = 1
	addrPort := netip.MustParseAddrPort("1.1.1.1:1234")
	cache := NewConnectionCache(1, 1*time.Minute, 1*time.Minute, "")
	cache.Set(id, addrPort)

	err := PersistEntriesToFile(persistancePath, cache)
	if err != nil {
		t.Fatalf("PersistEntriesToFile threw err: %v", err)
	}

	cache = NewConnectionCache(1, 1*time.Minute, 1*time.Minute, persistancePath)
	entry, ok := cache.entries[addrPort]
	if !ok {
		t.Fatal("entry not found in cache")
	}
	if entry.ID != id {
		t.Fatalf("entry id = %v; want %v", entry.ID, id)
	}

	err = os.Remove(persistancePath)
	if !ok {
		t.Log("failed to remove persistance file", err)
	}
}

func newConnectionCacheWithEntries(entrycount int) *ConnectionCache {
	cache := NewConnectionCache(1, 1*time.Minute, 1*time.Minute, "")
	rand.Seed(time.Now().UnixNano())
	var buf [4]byte

	for n := 0; n < entrycount; n++ {
		binary.LittleEndian.PutUint32(buf[:], rand.Uint32())

		cache.Set(int64(n), netip.AddrPortFrom(netip.AddrFrom4(buf), uint16(rand.Int31())))
	}

	return cache
}

func BenchmarkEncodeCacheEntries(b *testing.B) {
	for i := 1; i < 1e7; i *= 10 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			cache := newConnectionCacheWithEntries(i)

			b.ResetTimer()
			_, err := encodeCacheEntries(cache)
			b.StopTimer()

			if err != nil {
				b.Error("encodeCacheEntries threw error", err)
			}
		})
	}
}

func BenchmarkDecodeCacheEntries(b *testing.B) {
	b.StopTimer()
	for i := 1; i < 1e7; i *= 10 {
		b.Run(fmt.Sprintf("%d", i), func(b *testing.B) {
			cache := newConnectionCacheWithEntries(i)
			data, err := encodeCacheEntries(cache)
			if err != nil {
				b.Error("encodeCacheEntries threw error", err)
			}

			b.ResetTimer()
			_, err = decodeCacheEntries(data)
			b.StopTimer()

			if err != nil {
				b.Error("decodeCacheEntries threw error", err)
			}
		})
	}
}
