package cmd

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
		Debug:    struct{ Pprof int }{},
		Stats: struct {
			General  bool
			IP       bool
			Interval time.Duration
		}{},
		UDP: struct {
			Port        int
			IP          string
			Routines    int
			Connections struct {
				Validate bool
				GC       time.Duration
				Expiry   time.Duration
			}
		}{
			Connections: struct {
				Validate bool
				GC       time.Duration
				Expiry   time.Duration
			}{
				Validate: true,
				GC:       1 * time.Hour,
				Expiry:   1 * time.Hour,
			},
			Port:     1337,
			Routines: 1,
		},
		Announce: struct {
			Base time.Duration
			Fuzz time.Duration
		}{
			Fuzz: 1 * time.Second,
		},
		HTTP: struct {
			Tracker  bool
			Port     int
			IP       string
			Routines int
			Timeout  struct {
				Read  time.Duration
				Write time.Duration
			}
			Serve string
		}{
			Tracker:  true,
			Port:     1337,
			Routines: 1,
			Timeout: struct {
				Read  time.Duration
				Write time.Duration
			}{
				Read:  2 * time.Second,
				Write: 2 * time.Second,
			},
		},
		Numwant: struct {
			Default uint
			Limit   uint
		}{
			Default: 100,
			Limit:   100,
		},
		DB: struct {
			GC     time.Duration
			Expiry time.Duration
			Backup struct {
				Interval time.Duration
				Path     string
			}
		}{
			GC:     1 * time.Hour,
			Expiry: 1 * time.Hour,
		},
	}

	fmt.Println("Starting mock tracker...")
	go Run(conf)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("started!")

	m.Run()

	fmt.Println("Shutting down mock tracker...")
	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
}
