package storage

import "sync"

type Peerlist struct {
	Peers []byte
}

var peerlistPool sync.Pool

func GetPeerList() *Peerlist {
	return peerlistPool.Get().(*Peerlist)
}

func (p *Peerlist) Put() {
	p.Peers = p.Peers[:0]
	peerlistPool.Put(p)
}
