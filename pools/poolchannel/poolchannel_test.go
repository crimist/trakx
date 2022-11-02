package tmp

import (
	"sync"
	"testing"

	"github.com/crimist/trakx/tracker/storage"
)

const (
	channelMax    = 500_000
	channelBuffer = 5000
)

// --- PoolChannel parallel

func BenchmarkPoolChanelGet(b *testing.B) {
	b.ReportAllocs()
	poolChan := NewPoolChannel[storage.Peer](channelMax, channelBuffer)
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = poolChan.Get()
		}
	})
}

func BenchmarkPoolChanelGetPut(b *testing.B) {
	b.ReportAllocs()
	poolChan := NewPoolChannel[storage.Peer](channelMax, channelBuffer)
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			data := poolChan.Get()
			poolChan.Put(data)
		}
	})
}

// --- Pool parallel

func BenchmarkPoolGet(b *testing.B) {
	b.ReportAllocs()
	pool := sync.Pool{New: func() any {
		return new(storage.Peer)
	}}
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = pool.Get().(*storage.Peer)
		}
	})
}

func BenchmarkPoolGetPut(b *testing.B) {
	b.ReportAllocs()
	pool := sync.Pool{New: func() any {
		return new(storage.Peer)
	}}
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			data := pool.Get().(*storage.Peer)
			pool.Put(data)
		}
	})
}

// --- PoolChannel seq

func BenchmarkPoolChanelGetSeq(b *testing.B) {
	b.ReportAllocs()
	poolChan := NewPoolChannel[storage.Peer](channelMax, channelBuffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = poolChan.Get()
	}
}

func BenchmarkPoolChanelGetPutSeq(b *testing.B) {
	b.ReportAllocs()
	poolChan := NewPoolChannel[storage.Peer](channelMax, channelBuffer)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data := poolChan.Get()
		poolChan.Put(data)
	}
}

// --- Pool seq

func BenchmarkPoolGetSeq(b *testing.B) {
	b.ReportAllocs()
	pool := sync.Pool{New: func() any {
		return new(storage.Peer)
	}}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pool.Get().(*storage.Peer)
	}
}

func BenchmarkPoolGetPutSeq(b *testing.B) {
	b.ReportAllocs()
	pool := sync.Pool{New: func() any {
		return new(storage.Peer)
	}}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data := pool.Get().(*storage.Peer)
		pool.Put(data)
	}
}
