package shared

import (
	"time"

	"go.uber.org/zap"
)

const (
	// Ports
	HTTPPort   = "1337"
	UDPPort    = 1337
	ExpvarPort = "1338"

	// Tracker
	peerDBFilename = "peers.db"
	ConnDBFilename = "conn.db"
	DefaultNumwant = 75
	MaxNumwant     = 400
	Bye            = "See you space cowboy..."

	// Tracker intervals/timeout
	cleanInterval         = 3 * time.Minute
	UDPTrimInterval       = 5 * time.Minute
	metricsInterval       = 30 * time.Minute
	writeDBInterval       = 5 * time.Minute
	CleanTimeout    int64 = AnnounceInterval * 2 // 40 min

	// Client intervals
	AnnounceInterval = 20 * 60 // 20 min

)

var (
	PeerDB         PeerDatabase
	Logger         *zap.Logger
	Env            Enviroment
	UDPCheckConnID bool
)
