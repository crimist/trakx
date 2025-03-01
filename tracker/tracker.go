package tracker

import "time"

type Tracker interface {
	Serve() error
	Shutdown()
	ConnectionCount() int
}

type TrackerConfig struct {
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
	// TCP read timeout
	ReadTimeout time.Duration
	// TCP write timeout
	WriteTimeout time.Duration
}
