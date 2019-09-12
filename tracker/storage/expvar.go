// +build !fast

package storage

import (
	"sync"
	"sync/atomic"
)

const (
	fast     = false
	IPMapCap = 100000
)

var (
	Expvar expvals
)

type expvarIPmap struct {
	mu sync.Mutex
	M  map[PeerIP]int16
}

func (e *expvarIPmap) Lock() {
	e.mu.Lock()
}

func (e *expvarIPmap) Unlock() {
	e.mu.Unlock()
}

func (e *expvarIPmap) Delete(ip PeerIP) {
	delete(e.M, ip)
}

func (e *expvarIPmap) Inc(ip PeerIP) {
	e.M[ip]++
}

func (e *expvarIPmap) Dec(ip PeerIP) {
	e.M[ip]--
}

func (e *expvarIPmap) Dead(ip PeerIP) (dead bool) {
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

func init() { Expvar.IPs.M = make(map[PeerIP]int16, IPMapCap) }
