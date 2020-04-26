// +build fast

package storage

import (
	"sync"
)

const fast = true

type fakeExpvarInt struct{}

func (e fakeExpvarInt) Set(n int64)  {}
func (e fakeExpvarInt) Add(n int64)  {}
func (e fakeExpvarInt) Value() int64 { return -1 }

type expvals struct {
	IPs          expvarIPMap
	Hits         fakeExpvarInt
	Connects     fakeExpvarInt
	ConnectsOK   fakeExpvarInt
	Announces    fakeExpvarInt
	AnnouncesOK  fakeExpvarInt
	Scrapes      fakeExpvarInt
	ScrapesOK    fakeExpvarInt
	Errors       fakeExpvarInt
	ClientErrors fakeExpvarInt
	Seeds        fakeExpvarInt
	Leeches      fakeExpvarInt
	Pools        struct {
		Dict     fakeExpvarInt
		Peerlist fakeExpvarInt
		Peer     fakeExpvarInt
	}
}

var (
	Expvar expvals
)

type expvarIPMap struct {
	sync.Mutex
	submap map[PeerIP]int16
}

func (ipmap *expvarIPMap) Len() int         { return -1 }
func (ipmap *expvarIPMap) Delete(ip PeerIP) {}
func (ipmap *expvarIPMap) Inc(ip PeerIP)    {}
func (ipmap *expvarIPMap) Dec(ip PeerIP)    {}
func (ipmap *expvarIPMap) Remove(ip PeerIP) {}
