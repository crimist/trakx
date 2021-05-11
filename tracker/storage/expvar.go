// +build !fast

package storage

import (
	"expvar"
	"sync"
)

const (
	fast       = false
	ipmapAlloc = 200_000
)

type expvals struct {
	IPs          expvarIPMap
	Hits         *expvar.Int
	Connects     *expvar.Int
	ConnectsOK   *expvar.Int
	Announces    *expvar.Int
	AnnouncesOK  *expvar.Int
	Scrapes      *expvar.Int
	ScrapesOK    *expvar.Int
	Errors       *expvar.Int
	ClientErrors *expvar.Int
	Seeds        *expvar.Int
	Leeches      *expvar.Int
	Pools        struct {
		Dict     *expvar.Int
		Peerlist *expvar.Int
		Peer     *expvar.Int
	}
}

var (
	// Expvar contains the global expvars
	Expvar expvals
)

func init() {
	Expvar.IPs.submap = make(map[PeerIP]int16, ipmapAlloc)
	Expvar.Hits = expvar.NewInt("trakx.performance.hits")
	Expvar.Connects = expvar.NewInt("trakx.performance.connects")
	Expvar.ConnectsOK = expvar.NewInt("trakx.performance.connectsok")
	Expvar.Announces = expvar.NewInt("trakx.performance.announces")
	Expvar.AnnouncesOK = expvar.NewInt("trakx.performance.announcesok")
	Expvar.Scrapes = expvar.NewInt("trakx.performance.scrapes")
	Expvar.ScrapesOK = expvar.NewInt("trakx.performance.scrapesok")
	Expvar.Errors = expvar.NewInt("trakx.performance.errors")
	Expvar.ClientErrors = expvar.NewInt("trakx.performance.clienterrors")
	Expvar.Seeds = expvar.NewInt("trakx.database.seeds")
	Expvar.Leeches = expvar.NewInt("trakx.database.leeches")
	Expvar.Pools.Dict = expvar.NewInt("trakx.pools.dict")
	Expvar.Pools.Peerlist = expvar.NewInt("trakx.pools.peerlist")
	Expvar.Pools.Peer = expvar.NewInt("trakx.pools.peer")
}

type expvarIPMap struct {
	sync.Mutex
	submap map[PeerIP]int16
}

func (ipmap *expvarIPMap) Len() int {
	return len(ipmap.submap)
}

func (ipmap *expvarIPMap) Delete(ip PeerIP) {
	delete(ipmap.submap, ip)
}

func (ipmap *expvarIPMap) Inc(ip PeerIP) {
	ipmap.submap[ip]++
}

func (ipmap *expvarIPMap) Dec(ip PeerIP) {
	ipmap.submap[ip]--
}

// Remove decrements the IP and removes it if it's dead
func (ipmap *expvarIPMap) Remove(ip PeerIP) {
	ipmap.submap[ip]--
	if ipmap.submap[ip] < 1 {
		delete(ipmap.submap, ip)
	}
}
