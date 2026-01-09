package database

import (
	"io"
	"time"

	"github.com/crimist/trakx/stats"
)

type Config struct {
	InitalSize        int             // preallocted number of peers in the database
	EvictionFrequency time.Duration   // eviction frequency, 0 for no eviction
	ExpirationTime    time.Duration   // expiration time, ignored if eviction frequency is 0
	Collector         stats.Collector // statistics collector
	ImportReader      io.Reader       // optional snapshot reader to restore from on startup
}
