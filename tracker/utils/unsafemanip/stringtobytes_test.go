package unsafemanip

import (
	"strings"
	"testing"
)

var (
	d1      = strings.Repeat("test", 1e3)
	d1Bytes = []byte(d1)
	d2      = strings.Repeat("xxxx", 1e3)
	d2Bytes = []byte(d2)
)

// append is faster with same mem usage - use append where pre casting isn't required

func BenchmarkStringToBytesAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = StringToBytes(d1 + d2)
	}
}

func BenchmarkAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = append(d1Bytes, d2Bytes...)
	}
}
