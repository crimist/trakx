package experimentation

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/cornelk/hashmap"
)

// go test -v

const size = 40000

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

	fmt.Println("Heap Virtual Address Space:", now.HeapSys-then.HeapSys)

	fmt.Println("Number of GCs:", now.NumGC-then.NumGC)
}

func TestMemMap(t *testing.T) {
	start := GetMem()

	m := struct {
		sync.RWMutex
		m map[PeerID]*Peer
	}{m: make(map[PeerID]*Peer)}

	var id PeerID
	for i := 0; i < size; i++ {
		rand.Read(id[:])

		p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
		m.m[id] = &p
	}

	CalcMem(start)
}

func TestMemHashap(t *testing.T) {
	start := GetMem()

	m := &hashmap.HashMap{}

	var id PeerID
	for i := 0; i < size; i++ {
		rand.Read(id[:])

		p := Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
		m.Set(id[:], &p)
	}

	CalcMem(start)
}
