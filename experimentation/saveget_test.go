package experimentation

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/cornelk/hashmap"
	"github.com/patrickmn/go-cache"
)

// go test -bench=. -benchmem -run=^$

func BenchmarkHashmapVal(b *testing.B) {
	h := PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
	m := &hashmap.HashMap{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetHashedKey(uintptr(unsafe.Pointer(&h)), p)
	}
}

func BenchmarkHashmapPtr(b *testing.B) {
	h := PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
	m := &hashmap.HashMap{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.SetHashedKey(uintptr(unsafe.Pointer(&h)), &p)
		_, _ = m.GetHashedKey(uintptr(unsafe.Pointer(&h)))
	}
}

func BenchmarkMapVal(b *testing.B) {
	h := PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
	m := struct {
		sync.RWMutex
		m map[PeerID]Peer
	}{m: make(map[PeerID]Peer)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lock()
		m.m[h] = p
		_ = m.m[h]
		m.Unlock()
	}
}

func BenchmarkMapPtr(b *testing.B) {
	h := PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
	m := struct {
		sync.RWMutex
		m map[PeerID]*Peer
	}{m: make(map[PeerID]*Peer)}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Lock()
		m.m[h] = &p
		_ = m.m[h]
		m.Unlock()
	}
}

func BenchmarkGoCachePtr(b *testing.B) {
	h := PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
	m := cache.New(cache.NoExpiration, cache.NoExpiration)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := string(h[:])
		m.SetDefault(s, &p)
		_, _ = m.Get(s)
	}
}

func BenchmarkGoCacheVal(b *testing.B) {
	h := PeerID{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
	m := cache.New(cache.NoExpiration, cache.NoExpiration)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := string(h[:])
		m.SetDefault(s, p)
		_, _ = m.Get(s)
	}
}
