// +build fast

package storage

import (
	"sync/atomic"
)

const fast = true

var (
	Expvar expvals
)

type expvarIPmap struct {
	M map[PeerIP]int16
}

func (e *expvarIPmap) Lock()               {}
func (e *expvarIPmap) Unlock()             {}
func (e *expvarIPmap) Delete(ip PeerIP)    {}
func (e *expvarIPmap) Inc(ip PeerIP)       {}
func (e *expvarIPmap) Dec(ip PeerIP)       {}
func (e *expvarIPmap) Dead(ip PeerIP) bool { return false }

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
