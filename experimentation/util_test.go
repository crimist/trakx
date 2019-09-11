package experimentation

import (
	"crypto/rand"
	"testing"
	"unsafe"
)

// go test -bench=Util -benchmem -run=^$

func BenchmarkUtilHashToString(b *testing.B) {
	var h Hash
	rand.Read(h[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = string(h[:])
	}
}

type ByteCast struct {
	Addr *Peer
	Len  int
	Cap  int
}

// over 13x slower if not inlined
func Peer2Byte(addr *Peer) []byte {
	return *(*[]byte)(unsafe.Pointer(&ByteCast{addr, 16, 16}))
}

func BenchmarkUtilByteHack(b *testing.B) {
	p := &Peer{true, PeerIP{1, 2, 3, 4}, 8999, 4782948907612}
	var peerb []byte
	_ = peerb

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// peerb = *(*[]byte)(unsafe.Pointer(&ByteCast{p, 16, 16}))
		peerb = Peer2Byte(p)
	}
}
