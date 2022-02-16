//go:build !heroku
// +build !heroku

// Trakx controller entrypoint

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/crimist/trakx/controller"
)

func printHelp() {
	help := "Commands:\n"
	help += fmt.Sprintf("   %-12s returns the status of trakx\n", "status")
	help += fmt.Sprintf("   %-12s starts trakx as a service\n", "start")
	help += fmt.Sprintf("   %-12s stops trakx service\n", "stop")
	help += fmt.Sprintf("   %-12s restarts trakx service\n", "restart")
	help += fmt.Sprintf("   %-12s automatically starts trakx if it stops (doesn't return)\n", "watch")
	help += fmt.Sprintf("   %-12s runs trakx in shell (doesn't return)\n", "run")
	help += fmt.Sprintf("   %-12s wipes trakx pid file - use if you encounter errors with start/stop/restart commands\n", "reset")

	help += "Usage:\n"
	help += fmt.Sprintf("   %s <command>\n", os.Args[0])

	help += "Example:\n"
	help += fmt.Sprintf("   %s status # 'trakx is not running'\n", os.Args[0])

	fmt.Print(help)
}

func main() {
	writeFatal := func(err error) {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(-1)
	}

	if len(os.Args) < 2 {
		printHelp()
		return
	}

	c := controller.NewController()

	switch os.Args[1] {
	case "status":
		if c.Running() {
			fmt.Println("trakx is running")
		} else {
			fmt.Println("trakx is not running")
		}
	case "watch":
		for {
			if !c.Running() {
				if err := c.Start(); err != nil {
					writeFatal(err)
				}

				// Wait to let it set up
				time.Sleep(5 * time.Second)
			}
			time.Sleep(3 * time.Second)
		}
	case "run":
		c.Run()
	case "start":
		if err := c.Start(); err != nil {
			writeFatal(err)
		}
	case "stop":
		if err := c.Stop(); err != nil {
			writeFatal(err)
		}
	case "restart", "reboot":
		fmt.Println("rebooting...")
		if err := c.Stop(); err != nil {
			writeFatal(err)
		}
		if err := c.Start(); err != nil {
			writeFatal(err)
		}
		fmt.Println("rebooted!")
	case "reset":
		fmt.Println("wiping pid file...")
		if err := c.Wipe(); err != nil {
			writeFatal(err)
		}
		fmt.Println("wiped!")
	default:
		fmt.Fprintf(os.Stderr, "invalid command: '%s'\n\n", os.Args[1])
		printHelp()
	}
}
