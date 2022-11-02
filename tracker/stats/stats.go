package stats

import (
	"net/netip"
	"sync"
	"sync/atomic"
)

type ipStats struct {
	sync.Mutex
	submap map[netip.Addr]int16
}

var (
	// requests
	Hits      atomic.Int64 // requests received
	Connects  atomic.Int64 // udp connects
	Announces atomic.Int64 // announces
	Scrapes   atomic.Int64 // scrapes

	// db
	Seeds   atomic.Int64 // total seeds
	Leeches atomic.Int64 // total leeches
	IPStats ipStats      // total (unique) ips

	// errors
	ServerErrors atomic.Int64
	ClientErrors atomic.Int64
)
