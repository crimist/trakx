// +build !fast

package storage

import (
	"sync"
	"sync/atomic"
)

const (
	fast       = false
	IPMapAlloc = 100000
)

var (
	Expvar expvals
)

type expvarIPmap struct {
	sync.Mutex
	submap map[PeerIP]int16
}

func (e *expvarIPmap) Len() int {
	return len(e.submap)
}

func (e *expvarIPmap) Delete(ip PeerIP) {
	delete(e.submap, ip)
}

func (e *expvarIPmap) Inc(ip PeerIP) {
	e.submap[ip]++
}

func (e *expvarIPmap) Dec(ip PeerIP) {
	e.submap[ip]--
}

// Remove decrements the IP and removes it if it's dead
func (e *expvarIPmap) Remove(ip PeerIP) {
	e.submap[ip]--
	if e.submap[ip] < 1 {
		delete(e.submap, ip)
	}
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

func init() { Expvar.IPs.submap = make(map[PeerIP]int16, IPMapAlloc) }
