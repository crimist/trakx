package shared

import "time"

const (
	HTTPPort                 = "1337"
	UDPPort                  = 1337
	UDPTrimInterval          = 5 * time.Minute
	ExpvarPort               = "1338"
	AnnounceInterval         = 20 * 60                // 20 min
	CleanTimeout       int64 = AnnounceInterval + 120 // 22 min
	CleanInterval            = 3 * time.Minute
	WriteDBInterval          = 5 * time.Minute
	PeerDBFilename           = "trakx.db"
	PeerDBTempFilename       = "trakx.db.tmp"
	DefaultNumwant           = 75
	MaxNumwant               = 400
	Bye                      = "See you space cowboy..."
)
