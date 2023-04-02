package stats

import (
	"sync/atomic"
)

type Statistics struct {
	Hits      atomic.Int64 // requests received
	Connects  atomic.Int64 // udp connects
	Announces atomic.Int64 // announces
	Scrapes   atomic.Int64 // scrapes

	// db
	Seeds   atomic.Int64 // total seeds
	Leeches atomic.Int64 // total leeches
	IPStats ipStatistics // total (unique) ips

	// errors
	ServerErrors atomic.Int64
	ClientErrors atomic.Int64
}

func NewStats(ipPreallocate int) *Statistics {
	return &Statistics{
		IPStats: newIPStatistics(ipPreallocate),
	}
}
