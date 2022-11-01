package gomap

import (
	"runtime"
	"runtime/debug"
	"testing"
)

type encoder func() ([]byte, error)

func benchmarkMemuse(function encoder, b *testing.B) {
	b.StopTimer()
	b.ResetTimer()

	gcp := debug.SetGCPercent(-1)

	for i := 0; i < b.N; i++ {
		var start, end runtime.MemStats
		runtime.ReadMemStats(&start)

		b.StartTimer()
		encoded, _ := function()
		b.StopTimer()

		runtime.ReadMemStats(&end)

		b.Logf("Encode: %dMB using %dMB with %d GC cycles", len(encoded)/1024/1024, (end.HeapAlloc-start.HeapAlloc)/1024/1024, end.NumGC-start.NumGC)
		debug.FreeOSMemory()
	}

	debug.SetGCPercent(gcp)
}
