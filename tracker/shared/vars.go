package shared

import (
	"time"

	"go.uber.org/zap"
)

const (
	HTTPPort               = "1337"
	UDPPort                = 1337
	UDPTrimInterval        = 5 * time.Minute
	ExpvarPort             = "1338"
	AnnounceInterval       = 20 * 60              // 20 min
	CleanTimeout     int64 = AnnounceInterval * 2 // 40 min
	CleanInterval          = 3 * time.Minute
	WriteDBInterval        = 5 * time.Minute
	PeerDBFilename         = "peers.db"
	ConnDBFilename         = "conn.db"
	DefaultNumwant         = 75
	MaxNumwant             = 400
	Bye                    = "See you space cowboy..."
)

var (
	PeerDB         PeerDatabase
	Logger         *zap.Logger
	Env            Enviroment
	UDPCheckConnID bool
)
