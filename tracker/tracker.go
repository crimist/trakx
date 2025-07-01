package tracker

import (
	"net"
	"time"
)

type Tracker interface {
	Serve(ip net.IP, port int, routines int) error
	Shutdown()
}

type TrackerConfig struct {
	// TODO: remove this and put it in connections.NewConnections
	// Validate UDP connections
	Validate bool
	// Defualt number of peers to send to client
	DefaultNumwant uint
	// Maximum number of peers to send to client
	MaximumNumwant uint
	// Interval between announces
	Interval uint
	// Interval variance which smooths load spikes
	// `Interval + random(-IntervalVariance, IntervalVariance)`
	IntervalVariance uint
	// TODO: move these to http.NewTracker
	// TCP read timeout
	ReadTimeout time.Duration
	// TCP write timeout
	WriteTimeout time.Duration
}
