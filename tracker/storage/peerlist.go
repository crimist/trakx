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

func GetPeerList() *Peerlist {
	return peerlistPool.Get().(*Peerlist)
}

func (p *Peerlist) Put() {
	p.Data = p.Data[:peerlistMax]
	peerlistPool.Put(p)
}
