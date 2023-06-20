package inmemory

import (
	"time"

	"github.com/crimist/trakx/stats"
)

type Config struct {
	InitalSize         int                 // preallocted number of peers in the database, 0 for default
	Persistance        PersistanceStrategy // persistance strategy, nil for no persistance
	PersistanceAddress string              // persistance address, ignored if persistance is nil
	EvictionFrequency  time.Duration       // eviction frequency, 0 for no eviction
	ExpirationTime     time.Duration       // expiration time, ignored if eviction frequency is 0
	Stats              *stats.Statistics   // statistics, nil for no statistics
}
