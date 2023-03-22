package tracker

import (
	"fmt"
	"syscall"
	"testing"
	"time"

	"github.com/crimist/trakx/config"
)

func TestMain(m *testing.M) {
	conf := &config.Configuration{
		LogLevel: "debug",
		Debug: struct{ Pprof int }{
			Pprof: 0,
		},
		ExpvarInterval: 0,
		UDP: struct {
			Enabled bool
			IP      string
			Port    int
			Threads int
			ConnDB  struct {
				Validate bool
				Size     uint64
				Trim     time.Duration
				Expiry   time.Duration
			}
		}{
			ConnDB: struct {
				Validate bool
				Size     uint64
				Trim     time.Duration
				Expiry   time.Duration
			}{
				Validate: true,
				Trim:     1 * time.Hour,
				Expiry:   1 * time.Hour,
			},
			Enabled: true,
			Port:    1337,
			Threads: 1,
		},
		Announce: struct {
			Base time.Duration
			Fuzz time.Duration
		}{
			Base: 0,
			Fuzz: 1 * time.Second,
		},
		HTTP: struct {
			Mode    string
			IP      string
			Port    int
			Timeout struct {
				Read  time.Duration
				Write time.Duration
			}
			Threads int
		}{
			Mode: "enabled",
			Port: 1337,
			Timeout: struct {
				Read  time.Duration
				Write time.Duration
			}{
				Read:  2 * time.Second,
				Write: 2 * time.Second,
			},
			Threads: 1,
		},
		Numwant: struct {
			Default uint
			Limit   uint
		}{
			Default: 100,
			Limit:   100,
		},
		DB: struct {
			Type   string
			Backup struct {
				Frequency time.Duration
				Type      string
				Path      string
			}
			Trim   time.Duration
			Expiry time.Duration
		}{
			Type:   "gomap",
			Trim:   1 * time.Hour,
			Expiry: 1 * time.Hour,
			Backup: struct {
				Frequency time.Duration
				Type      string
				Path      string
			}{
				Type:      "none",
				Frequency: 0,
			},
		},
	}

	// run tracker
	fmt.Println("Starting mock tracker...")
	go Run(conf)
	time.Sleep(100 * time.Millisecond) // wait for run to complete
	fmt.Println("started!")

	m.Run()

	fmt.Println("Shutting down mock tracker...")
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
}
