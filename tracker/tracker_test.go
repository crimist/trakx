package tracker

import (
	"fmt"
	"time"

	"github.com/crimist/trakx/tracker/config"
)

func init() {
	oneHour := 1 * time.Hour

	// mock config
	config.Conf.LogLevel = "debug"

	config.Conf.Debug.Pprof = 0
	config.Conf.ExpvarInterval = 0
	config.Conf.Debug.NofileLimit = 0
	config.Conf.DB.PeerPointers = 0
	config.Conf.UDP.ConnDB.Validate = true

	config.Conf.Announce.Base = 0
	config.Conf.Announce.Fuzz = 1 * time.Second
	config.Conf.HTTP.Mode = "enabled"
	config.Conf.HTTP.Port = 1337
	config.Conf.HTTP.Timeout.Read = 2
	config.Conf.HTTP.Timeout.Write = 2
	config.Conf.HTTP.Threads = 1
	config.Conf.UDP.Enabled = true
	config.Conf.UDP.Port = 1337
	config.Conf.UDP.Threads = 1
	config.Conf.Numwant.Default = 100
	config.Conf.Numwant.Limit = 100

	config.Conf.DB.Type = "gomap"
	config.Conf.DB.Backup.Type = "none"
	config.Conf.DB.Trim = oneHour
	config.Conf.DB.Backup.Frequency = 0
	config.Conf.DB.Expiry = oneHour
	config.Conf.UDP.ConnDB.Trim = oneHour
	config.Conf.UDP.ConnDB.Expiry = oneHour

	// run tracker
	fmt.Print("Starting mock tracker... ")
	go Run()
	time.Sleep(1000 * time.Millisecond) // wait for run to complete
	fmt.Println("started!")
}
