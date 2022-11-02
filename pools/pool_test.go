package pools

import (
	"sync"
	"testing"

	"github.com/crimist/trakx/bencoding"
)

// --- sync.Pool

func BenchmarkSyncPoolGet(b *testing.B) {
	b.ReportAllocs()

	pool := sync.Pool{
		New: func() any {
			return bencoding.NewDictionary()
		},
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = pool.Get().(*bencoding.Dictionary)
		}
	})
}

func BenchmarkSyncPoolGetPut(b *testing.B) {
	b.ReportAllocs()

	pool := sync.Pool{
		New: func() any {
			return bencoding.NewDictionary()
		},
	}

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			data := pool.Get().(*bencoding.Dictionary)
			data.Reset()
			pool.Put(data)
		}
	})
}

// --- custom pool

func BenchmarkPoolGet(b *testing.B) {
	b.ReportAllocs()

	pool := NewPool(func() any {
		return bencoding.NewDictionary()
	}, func(dictionary *bencoding.Dictionary) {
		dictionary.Reset()
	})

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			_ = pool.Get()
		}
	})

	b.Logf("created = %v", pool.Created())
}

func BenchmarkPoolGetPut(b *testing.B) {
	b.ReportAllocs()

	pool := NewPool(func() any {
		return bencoding.NewDictionary()
	}, func(dictionary *bencoding.Dictionary) {
		dictionary.Reset()
	})

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			data := pool.Get()
			pool.Put(data)
		}
	})

	b.Logf("created = %v", pool.Created())
}
