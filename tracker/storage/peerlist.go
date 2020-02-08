package storage

import (
	"sync"

	"github.com/crimist/trakx/tracker/shared"
)

type Peerlist struct {
	Data []byte
}

var peerlistMax = new(int)
var peerlistPool = sync.Pool{New: func() interface{} {
	p := new(Peerlist)
	p.Data = make([]byte, *peerlistMax)
	return p
}}

func GetPeerList() *Peerlist {
	return peerlistPool.Get().(*Peerlist)
}

func (p *Peerlist) Put() {
	shared.SetSliceLen(&p.Data, *peerlistMax)
	peerlistPool.Put(p)
}
