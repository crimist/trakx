package experimentation

import (
	"fmt"
	"runtime"
	"testing"
)

// go test -v

const size = 10000

func GetMem() runtime.MemStats {
	var now runtime.MemStats
	runtime.ReadMemStats(&now)
	return now
}

func CalcMem(then runtime.MemStats) {
	var now runtime.MemStats
	runtime.ReadMemStats(&now)

	// Add more: https://golang.org/pkg/runtime/#MemStats

	fmt.Println("Heap Bytes:", now.Alloc-then.Alloc)
	fmt.Println("Heap Mallocs:", now.Mallocs-then.Mallocs)
	fmt.Println("Heap Frees:", now.Frees-then.Frees)

	fmt.Println("Heap Virtual Address Space:", now.HeapSys-then.HeapSys)

	fmt.Println("Number of GCs:", now.NumGC-then.NumGC)
}

func TestMemMap(t *testing.T) {
	start := GetMem()

	// make map and fill it

	CalcMem(start)
}
