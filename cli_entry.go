//go:build !heroku
// +build !heroku

// Trakx controller entrypoint

package main

import (
	"fmt"
	"os"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/controller"
	"go.uber.org/zap"
)

func printHelp() {
	help := "Commands:\n"
	help += fmt.Sprintf("   %-12s returns the status of trakx\n", "status")
	help += fmt.Sprintf("   %-12s starts trakx daemon\n", "start")
	help += fmt.Sprintf("   %-12s stops trakx daemon\n", "stop")
	help += fmt.Sprintf("   %-12s restarts trakx daemon\n", "restart")
	help += fmt.Sprintf("   %-12s executes trakx, doesn't return\n", "execute")
	help += fmt.Sprintf("   %-12s wipes trakx pid file, use if you encounter errors with start/stop/restart commands\n", "reset")

	help += "Usage:\n"
	help += fmt.Sprintf("   %s <command>\n", os.Args[0])

	help += "Example:\n"
	help += fmt.Sprintf("   %s status\n", os.Args[0])

	fmt.Print(help)
}

func logFatal(err error) {
	fmt.Fprintf(os.Stderr, err.Error()+"\n")
	os.Exit(-1)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	conf, err := config.Load()
	if err != nil {
		zap.L().Fatal("failed to load configuration", zap.Error(err))
	}

	controller := controller.NewController(conf)

	switch os.Args[1] {
	case "status":
		pidFileExists, processAlive, heartbeat := controller.Status()

		if pidFileExists {
			fmt.Println("[✅] valid process id file")
		} else {
			fmt.Println("[❌] invalid/empty process id file")
		}
		if processAlive {
			fmt.Println("[✅] process is alive")
		} else {
			fmt.Println("[❌] process is dead")
		}
		if heartbeat {
			fmt.Println("[✅] heartbeat successful")
		} else {
			fmt.Println("[❌] heartbeat failed")
		}
	case "execute":
		controller.Execute()
	case "start":
		if err := controller.Start(); err != nil {
			logFatal(err)
		}
	case "stop":
		if err := controller.Stop(); err != nil {
			logFatal(err)
		}
	case "restart", "reboot":
		fmt.Println("rebooting...")
		if err := controller.Stop(); err != nil {
			logFatal(err)
		}
		if err := controller.Start(); err != nil {
			logFatal(err)
		}
		fmt.Println("rebooted!")
	case "reset":
		fmt.Println("wiping pid file...")
		if err := controller.Clear(); err != nil {
			logFatal(err)
		}
		fmt.Println("wiped!")
	default:
		fmt.Fprintf(os.Stderr, "invalid command: '%s'\n\n", os.Args[1])
		printHelp()
	}
}
