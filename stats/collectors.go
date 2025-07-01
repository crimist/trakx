package stats

import (
	"net/netip"
	"sync"
	"sync/atomic"
)

type atomicCollector struct {
	hits           atomic.Int64
	connects       atomic.Int64
	announces      atomic.Int64
	scrapes        atomic.Int64
	seeds          atomic.Int64
	leeches        atomic.Int64
	errorResponses atomic.Int64
	ipStats        IPCollector
}

func (s *atomicCollector) Hit()               { s.hits.Add(1) }
func (s *atomicCollector) Connect()           { s.connects.Add(1) }
func (s *atomicCollector) Announce()          { s.announces.Add(1) }
func (s *atomicCollector) Scrape()            { s.scrapes.Add(1) }
func (s *atomicCollector) AddSeeds(n int64)   { s.seeds.Add(n) }
func (s *atomicCollector) AddLeeches(n int64) { s.leeches.Add(n) }
func (s *atomicCollector) ErrorResponse()     { s.errorResponses.Add(1) }
func (s *atomicCollector) IPs() IPCollector   { return s.ipStats }
func (s *atomicCollector) GetSnapshot() Snapshot {
	return Snapshot{
		Hits:           s.hits.Load(),
		Connects:       s.connects.Load(),
		Announces:      s.announces.Load(),
		Scrapes:        s.scrapes.Load(),
		Seeds:          s.seeds.Load(),
		Leeches:        s.leeches.Load(),
		IPs:            s.ipStats.Total(),
		ErrorResponses: s.errorResponses.Load(),
	}
}

type mapIPCollector struct {
	sync.Mutex
	submap map[netip.Addr]int16
}

func (ipc *mapIPCollector) Inc(ip netip.Addr) {
	ipc.Lock()
	ipc.submap[ip]++
	ipc.Unlock()
}

func (ipc *mapIPCollector) Remove(ip netip.Addr) {
	ipc.Lock()
	ipc.submap[ip]--
	if ipc.submap[ip] == 0 {
		delete(ipc.submap, ip)
	}
	ipc.Unlock()
}

func (ipc *mapIPCollector) Replace(oldIP, newIP netip.Addr) {
	if oldIP == newIP {
		return
	}

	ipc.Lock()
	ipc.submap[oldIP]--
	if ipc.submap[oldIP] == 0 {
		delete(ipc.submap, oldIP)
	}
	ipc.submap[newIP]++
	ipc.Unlock()
}

func (ipc *mapIPCollector) Total() int {
	ipc.Lock()
	defer ipc.Unlock()
	return len(ipc.submap)
}
