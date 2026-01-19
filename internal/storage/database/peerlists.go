package database

import "github.com/crimist/trakx/internal/pool"

const (
	peerlist4Stride = 6  // ipv4 (4 bytes) + port (2 bytes)
	peerlist6Stride = 18 // ipv6 (16 bytes) + port (2 bytes)
)

type peerListPool struct {
	v4    *pool.Pool[[]byte]
	v6    *pool.Pool[[]byte]
	size4 int
	size6 int
}

func newPeerListPool(maxNumwant uint) *peerListPool {
	if maxNumwant == 0 {
		return nil
	}

	size4 := int(maxNumwant) * peerlist4Stride
	size6 := int(maxNumwant) * peerlist6Stride

	return &peerListPool{
		size4: size4,
		size6: size6,
		v4: pool.New(func() []byte {
			return make([]byte, size4)
		}, func(b []byte) []byte {
			return b[:size4]
		}),
		v6: pool.New(func() []byte {
			return make([]byte, size6)
		}, func(b []byte) []byte {
			return b[:size6]
		}),
	}
}

func (p *peerListPool) getV4() []byte {
	if p == nil {
		return nil
	}
	return p.v4.Get()
}

func (p *peerListPool) getV6() []byte {
	if p == nil {
		return nil
	}
	return p.v6.Get()
}

func (p *peerListPool) putV4(b []byte) {
	if p == nil || b == nil {
		return
	}
	p.v4.Put(b)
}

func (p *peerListPool) putV6(b []byte) {
	if p == nil || b == nil {
		return
	}
	p.v6.Put(b)
}

func (p *peerListPool) stats() (int64, int64) {
	if p == nil {
		return 0, 0
	}
	return p.v4.Created(), p.v6.Created()
}
