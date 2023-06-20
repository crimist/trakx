package stats

import (
	"net/netip"
	"sync"
)

type ipStatistics struct {
	sync.Mutex
	ips map[netip.Addr]int16
}

func newIPStatistics(prealloc int) ipStatistics {
	return ipStatistics{
		ips: make(map[netip.Addr]int16, prealloc),
	}
}
