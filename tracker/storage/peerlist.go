package storage

import (
	"sync"
)

type Peerlist struct {
	Data []byte
}

var peerlistMax int

var peerlistPool = sync.Pool{New: func() interface{} {
	Expvar.Pools.Peerlist.Add(1)
	p := new(Peerlist)
	p.Data = make([]byte, peerlistMax)
	return p
}}

// GetPeerList returns a peerlist pointer for use.
func GetPeerList() *Peerlist {
	return peerlistPool.Get().(*Peerlist)
}

// Put clears and returns a Peerlist to the peerlistPool after use.
func (p *Peerlist) Put() {
	p.Data = p.Data[:peerlistMax]
	peerlistPool.Put(p)
}
