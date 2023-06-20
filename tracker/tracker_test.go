package tracker

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/crimist/trakx/config"
)

func TestMain(m *testing.M) {
	oneHour := 1 * time.Hour

	// mock config
	config.Config.LogLevel = "debug"

	config.Config.Debug.Pprof = 0
	config.Config.ExpvarInterval = 0
	config.Config.UDP.ConnDB.Validate = true

	config.Config.Announce.Base = 0
	config.Config.Announce.Fuzz = 1 * time.Second
	config.Config.HTTP.Mode = "enabled"
	config.Config.HTTP.Port = 1337
	config.Config.HTTP.Timeout.Read = 2
	config.Config.HTTP.Timeout.Write = 2
	config.Config.HTTP.Threads = 1
	config.Config.UDP.Enabled = true
	config.Config.UDP.Port = 1337
	config.Config.UDP.Threads = 1
	config.Config.Numwant.Default = 100
	config.Config.Numwant.Limit = 100

	config.Config.DB.Type = "gomap"
	config.Config.DB.Backup.Type = "none"
	config.Config.DB.Trim = oneHour
	config.Config.DB.Backup.Frequency = 0
	config.Config.DB.Expiry = oneHour
	config.Config.UDP.ConnDB.Trim = oneHour
	config.Config.UDP.ConnDB.Expiry = oneHour

	// run tracker
	fmt.Println("Starting mock tracker...")
	go Run()
	time.Sleep(100 * time.Millisecond) // wait for run to complete
	fmt.Println("started!")

	m.Run()

	fmt.Println("Shutting down mock tracker...")
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
}
