// +build expvar

package shared

import (
	"sync"
	"sync/atomic"
)

const expvarOn = true

var (
	Expvar expvals
)

type expvarIPmap struct {
	mu sync.Mutex
	M  map[PeerIP]int8
}

func (e *expvarIPmap) Lock() {
	e.mu.Lock()
}

func (e *expvarIPmap) Unlock() {
	e.mu.Unlock()
}

func (e *expvarIPmap) delete(ip PeerIP) {
	delete(e.M, ip)
}

func (e *expvarIPmap) inc(ip PeerIP) {
	e.M[ip]++
}

func (e *expvarIPmap) dec(ip PeerIP) {
	e.M[ip]--
}

func (e *expvarIPmap) dead(ip PeerIP) (dead bool) {
	if e.M[ip] < 1 {
		dead = true
	}
	return
}

type expvals struct {
	Connects    int64
	ConnectsOK  int64
	Announces   int64
	AnnouncesOK int64
	Scrapes     int64
	ScrapesOK   int64
	Errs        int64
	Clienterrs  int64
	Seeds       int64
	Leeches     int64
	IPs         expvarIPmap
}

func AddExpval(num *int64, inc int64) {
	atomic.AddInt64(num, inc)
}

// InitExpvar sets the expvar vars to the contents of the peer database
func InitExpvar(peerdb *PeerDatabase) {
	if ok := peerdb.check(); !ok {
		panic("peerdb not init before expvars")
	}

	Expvar.IPs.M = make(map[PeerIP]int8, 30000)

	Expvar.IPs.Lock()
	for _, peermap := range peerdb.db {
		for _, peer := range peermap {
			Expvar.IPs.M[peer.IP]++
			if peer.Complete == true {
				Expvar.Seeds++
			} else {
				Expvar.Leeches++
			}
		}
	}
	Expvar.IPs.Unlock()
}
