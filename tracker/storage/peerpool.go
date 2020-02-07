package storage

import (
	"sync"
)

// this is probably gonna create some really hard to debug memory issues
var peerPool = sync.Pool{New: func() interface{} {
	println("New peer from pool")
	return new(Peer)
}}

func GetPeer() *Peer {
	return peerPool.Get().(*Peer)
}

func (p *Peer) Put() {
	peerPool.Put(p)
}
