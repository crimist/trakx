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
	help += fmt.Sprintf("   %-12s executes trakx in shell (doesn't return)\n", "execute")
	help += fmt.Sprintf("   %-12s wipes trakx pid file - use if you encounter errors with start/stop/restart commands\n", "reset")

	help += "Usage:\n"
	help += fmt.Sprintf("   %s <command>\n", os.Args[0])

	help += "Example:\n"
	help += fmt.Sprintf("   %s status # 'trakx is not running'\n", os.Args[0])

	fmt.Print(help)
}

func logErrorFatal(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
	os.Exit(-1)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	controller := controller.NewController()

	switch os.Args[1] {
	case "status":
		pidFileExists, processAlive, heartbeat := controller.Status()

		if pidFileExists {
			fmt.Println("[OK] read valid process id file")
		} else {
			fmt.Println("[Err] failed to read process id file")
		}
		if processAlive {
			fmt.Println("[OK] process is alive")
		} else {
			fmt.Println("[Err] process is dead")
		}
		if heartbeat {
			fmt.Println("[OK] heartbeat successful")
		} else {
			fmt.Println("[Err] heartbeat failed")
		}

		if pidFileExists || processAlive || heartbeat {
			fmt.Println("trakx is running!")
		}
	case "watch":
		for {
			pidFileExists, processAlive, heartbeat := controller.Status()

			if !(pidFileExists && processAlive && heartbeat) {
				if err := controller.Start(); err != nil {
					logErrorFatal(err)
				}
			}
			time.Sleep(5 * time.Second)
		}
	case "execute":
		controller.Execute()
	case "start":
		if err := controller.Start(); err != nil {
			logErrorFatal(err)
		}
	case "stop":
		if err := controller.Stop(); err != nil {
			logErrorFatal(err)
		}
	case "restart", "reboot":
		fmt.Println("rebooting...")
		if err := controller.Stop(); err != nil {
			logErrorFatal(err)
		}
		if err := controller.Start(); err != nil {
			logErrorFatal(err)
		}
		fmt.Println("rebooted!")
	case "reset":
		fmt.Println("wiping pid file...")
		if err := controller.Clear(); err != nil {
			logErrorFatal(err)
		}
		fmt.Println("wiped!")
	default:
		fmt.Fprintf(os.Stderr, "invalid command: '%s'\n\n", os.Args[1])
		printHelp()
	}
}
