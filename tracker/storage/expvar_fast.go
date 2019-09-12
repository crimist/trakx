// +build fast

package storage

import (
	"sync/atomic"

	"github.com/syc0x00/trakx/tracker/shared"
)

const fast = true

var (
	Expvar expvals
)

type expvarIPmap struct {
	M map[shared.PeerIP]int16
}

func (e *expvarIPmap) Lock()                      {}
func (e *expvarIPmap) Unlock()                    {}
func (e *expvarIPmap) Delete(ip shared.PeerIP)    {}
func (e *expvarIPmap) Inc(ip shared.PeerIP)       {}
func (e *expvarIPmap) Dec(ip shared.PeerIP)       {}
func (e *expvarIPmap) Dead(ip shared.PeerIP) bool { return false }

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
