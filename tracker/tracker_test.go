package tracker

import (
	"fmt"
	"math"
	"time"

	"github.com/crimist/trakx/tracker/config"
)

func init() {
	intMax := int(math.Pow(2, 32))
	int64Max := int64(math.Pow(2, 32))

	// mock config
	config.Conf.LogLevel = "debug"

	config.Conf.Debug.PprofPort = 0
	config.Conf.Debug.ExpvarInterval = 0
	config.Conf.Debug.NofileLimit = 0
	config.Conf.Debug.PeerChanInit = 0
	config.Conf.Debug.CheckConnIDs = true

	config.Conf.Tracker.Announce = 0
	config.Conf.Tracker.AnnounceFuzz = 1
	config.Conf.Tracker.HTTP.Mode = "enabled"
	config.Conf.Tracker.HTTP.Port = 1337
	config.Conf.Tracker.HTTP.ReadTimeout = 2
	config.Conf.Tracker.HTTP.WriteTimeout = 2
	config.Conf.Tracker.HTTP.Threads = 1
	config.Conf.Tracker.UDP.Enabled = true
	config.Conf.Tracker.UDP.Port = 1337
	config.Conf.Tracker.UDP.Threads = 1
	config.Conf.Tracker.Numwant.Default = 100
	config.Conf.Tracker.Numwant.Limit = 100

	config.Conf.Database.Type = "gomap"
	config.Conf.Database.Backup = "none"
	config.Conf.Database.Peer.Trim = intMax
	config.Conf.Database.Peer.Write = 0
	config.Conf.Database.Peer.Timeout = int64Max
	config.Conf.Database.Conn.Trim = intMax
	config.Conf.Database.Conn.Timeout = int64Max

	// run tracker
	fmt.Print("Starting mock tracker... ")
	go Run()
	time.Sleep(1000 * time.Millisecond) // wait for run to complete
	fmt.Println("started!")
}
