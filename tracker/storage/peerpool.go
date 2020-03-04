package storage

import (
	"sync"
)

// this is probably gonna create some really hard to debug memory issues
var peerPool = sync.Pool{New: func() interface{} {
	Expvar.Pools.Peer.Add(1)
	return new(Peer)
}}

func GetPeer() *Peer {
	return peerPool.Get().(*Peer)
}

func (p *Peer) Put() {
	peerPool.Put(p)
}
