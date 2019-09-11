package experimentation

import (
	"crypto/rand"
	"fmt"
	"log"
	mathrand "math/rand"
	"runtime"
	"runtime/debug"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/allegro/bigcache"
	"github.com/coocood/freecache"
	"github.com/cornelk/hashmap"
	"github.com/patrickmn/go-cache"
)

// go test -v

func init() {
	// disable GC
	debug.SetGCPercent(-1)
}

const (
	hashes = 100_000
	pph    = 2
)

func GetMem() runtime.MemStats {
	var now runtime.MemStats
	runtime.ReadMemStats(&now)
	return now
}

func CalcMem(then runtime.MemStats) {
	var now runtime.MemStats
	runtime.ReadMemStats(&now)

	// Add more: https://golang.org/pkg/runtime/#MemStats

	fmt.Println("Heap Bytes (mb):", float32(now.Alloc-then.Alloc)/1024.0/1024.0)
	fmt.Println("Heap Mallocs:", now.Mallocs-then.Mallocs)
	fmt.Println("Heap Frees:", now.Frees-then.Frees)

	if now.HeapSys > then.HeapSys {
		fmt.Println("Heap VAS:", now.HeapSys-then.HeapSys)
	} else {
		fmt.Println("Heap VAS:", "< 0")
	}

	fmt.Println("Number of GCs:", now.NumGC-then.NumGC)

	// runtime.GC()
	debug.FreeOSMemory()
}

type submap struct {
	sync.RWMutex
	m map[PeerID]*Peer
}

//go:linkname localname runtime/thunk.s
func TestMemMap(t *testing.T) {
	start := GetMem()

	m := struct {
		sync.RWMutex
		sub map[Hash]*submap
	}{sub: make(map[Hash]*submap, hashes)}

	for i := 0; i < hashes; i++ {
		var h Hash
		rand.Read(h[:])

		// init submap
		x := new(submap)
		x.m = make(map[PeerID]*Peer, 2)
		m.sub[h] = x

		for n := 0; n < pph; n++ {
			var id PeerID
			rand.Read(id[:])
			p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, mathrand.Int63()}

			m.sub[h].m[id] = &p
		}
	}

	CalcMem(start)
}

func TestMemHashap(t *testing.T) {
	t.Skip("Too slow")

	start := GetMem()

	m := hashmap.New(1)

	for i := 0; i < hashes; i++ {
		var h Hash
		rand.Read(h[:])

		// init submap
		sub := hashmap.New(1)
		m.SetHashedKey(uintptr(unsafe.Pointer(&h)), &sub)

		for n := 0; n < pph; n++ {
			var id PeerID
			rand.Read(id[:])
			p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, mathrand.Int63()}

			m.SetHashedKey(uintptr(unsafe.Pointer(&id)), &p)
		}
	}

	CalcMem(start)
}

func TestMemGoCache(t *testing.T) {
	start := GetMem()

	m := cache.New(cache.NoExpiration, cache.NoExpiration)

	for i := 0; i < hashes; i++ {
		var h Hash
		rand.Read(h[:])

		// init submap
		subcache := cache.New(5*time.Minute, 10*time.Minute)
		m.SetDefault(string(h[:]), subcache)

		for n := 0; n < pph; n++ {
			var id PeerID
			rand.Read(id[:])
			p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, mathrand.Int63()}

			subcache.SetDefault(string(id[:]), &p)
		}
	}

	CalcMem(start)
}

func TestMemFreecache(t *testing.T) {
	t.Skip("Far too big in memory")
	start := GetMem()

	m := make(map[Hash]*freecache.Cache)

	for i := 0; i < hashes; i++ {
		var h Hash
		rand.Read(h[:])

		// init submap
		subcache := freecache.NewCache(16 * 1024) // min size
		m[h] = subcache

		for n := 0; n < pph; n++ {
			var id PeerID
			rand.Read(id[:])
			p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, mathrand.Int63()}

			var peerb []byte = *(*[]byte)(unsafe.Pointer(&p))

			subcache.Set(id[:], peerb, 0)
		}
	}

	CalcMem(start)
}

func TestMemBigcache(t *testing.T) {
	start := GetMem()

	config := bigcache.Config{
		Shards:             2,
		LifeWindow:         0,
		CleanWindow:        0,
		MaxEntriesInWindow: 1,
		MaxEntrySize:       16,
		Verbose:            true,
		HardMaxCacheSize:   1024, // max in mb
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}

	m := make(map[Hash]*bigcache.BigCache)

	for i := 0; i < hashes; i++ {
		var h Hash
		rand.Read(h[:])

		// init submap
		subcache, initErr := bigcache.NewBigCache(config)
		if initErr != nil {
			log.Fatal(initErr)
		}
		m[h] = subcache

		for n := 0; n < pph; n++ {
			var id PeerID
			rand.Read(id[:])
			p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, mathrand.Int63()}

			subcache.Set(string(id[:]), Peer2Byte(&p))
		}
	}

	CalcMem(start)
}
