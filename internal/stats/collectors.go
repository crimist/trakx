package stats

import (
	"expvar"
	"net/netip"
	"sync"
)

type expvarCollector struct {
	hits      *expvar.Int
	connects  *expvar.Int
	announces *expvar.Int
	scrapes   *expvar.Int
	seeds     *expvar.Int
	leeches   *expvar.Int
	errors    *expvar.Int
	ipStats   IPCollector
}

func (s *expvarCollector) Hit()               { s.hits.Add(1) }
func (s *expvarCollector) Connect()           { s.connects.Add(1) }
func (s *expvarCollector) Announce()          { s.announces.Add(1) }
func (s *expvarCollector) Scrape()            { s.scrapes.Add(1) }
func (s *expvarCollector) AddSeeds(n int64)   { s.seeds.Add(n) }
func (s *expvarCollector) AddLeeches(n int64) { s.leeches.Add(n) }
func (s *expvarCollector) ErrorResponse()     { s.errors.Add(1) }
func (s *expvarCollector) IPs() IPCollector   { return s.ipStats }
func (s *expvarCollector) GetSnapshot() Snapshot {
	return Snapshot{
		Hits:      s.hits.Value(),
		Connects:  s.connects.Value(),
		Announces: s.announces.Value(),
		Scrapes:   s.scrapes.Value(),
		Seeds:     s.seeds.Value(),
		Leeches:   s.leeches.Value(),
		IPs:       s.ipStats.Total(),
		Errors:    s.errors.Value(),
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
