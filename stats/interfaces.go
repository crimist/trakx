package stats

import (
	"net/netip"
)

// Snapshot holds a point-in-time copy of the statistics.
type Snapshot struct {
	// Tracker requests received
	Hits int64
	// UDP connect requests
	Connects int64
	// Tracker announce requests
	Announces int64
	// Tracker scrape requests
	Scrapes int64

	Seeds   int64
	Leeches int64
	// Number of unique IPs in the database
	IPs int

	// Number of error responses from tracker
	ErrorResponses int64
}

// IPCollector defines the interface for collecting IP statistics.
type IPCollector interface {
	Inc(ip netip.Addr)
	Remove(ip netip.Addr)
	Replace(oldIP, newIP netip.Addr)
	Total() int
}

// Collector defines the interface for collecting statistics.
type Collector interface {
	Hit()
	Connect()
	Announce()
	Scrape()
	AddSeeds(n int64)
	AddLeeches(n int64)
	ErrorResponse()
	GetSnapshot() Snapshot
	IPs() IPCollector
}

func NewCollectors(enabled, ipStatsEnabled bool, ipMapPrealloc int) Collector {
	if !enabled {
		return &noopCollector{}
	}

	var ipCollector IPCollector
	if ipStatsEnabled {
		ipCollector = &mapIPCollector{
			submap: make(map[netip.Addr]int16, ipMapPrealloc),
		}
	} else {
		ipCollector = &noopIPCollector{}
	}

	return &atomicCollector{
		ipStats: ipCollector,
	}
}
