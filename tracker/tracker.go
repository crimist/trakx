package tracker

import (
	"net"
)

type Tracker interface {
	Serve(ip net.IP, port int, routines int) error
	Shutdown()
}

type TrackerConfig struct {
	// Defualt number of peers to send to client
	DefaultNumwant uint
	// Maximum number of peers to send to client
	MaximumNumwant uint
	// Interval between announces
	Interval uint
	// Interval variance which smooths load spikes
	// `Interval + random(-IntervalVariance, IntervalVariance)`
	IntervalVariance uint
}
