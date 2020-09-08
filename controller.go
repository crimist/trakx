// +build !heroku
// Trakx controller

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/crimist/trakx/controller"
)

func printHelp() {
	help := "Commands:\n"
	help += fmt.Sprintf("  %-12s Checks if Trakx is running\n", "status")
	help += fmt.Sprintf("  %-12s Watches if Trakx stops and starts it again if it does (doesn't return)\n", "watch")
	help += fmt.Sprintf("  %-12s Runs Trakx (doesn't return)\n", "run")
	help += fmt.Sprintf("  %-12s Starts Trakx as a service\n", "start")
	help += fmt.Sprintf("  %-12s Stops Trakx service\n", "stop")
	help += fmt.Sprintf("  %-12s Restarts Trakx service\n", "restart")
	help += fmt.Sprintf("  %-12s Wipes trakx pid file\n", "clean")

	help += "Usage:\n"
	help += fmt.Sprintf("  %s <command>\n", os.Args[0])

	help += "Example:\n"
	help += fmt.Sprintf("  %s status # \"Trakx is not running\"\n", os.Args[0])

	fmt.Print(help)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	c := controller.NewController()

	switch os.Args[1] {
	case "status":
		if c.Running() {
			fmt.Println("Trakx is running")
		} else {
			fmt.Println("Trakx is not running")
		}
	case "watch":
		for {
			if !c.Running() {
				if err := c.Start(); err != nil {
					fmt.Fprintf(os.Stderr, err.Error()+"\n")
					os.Exit(-1)
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
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
	case "stop":
		if err := c.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
	case "restart", "reboot":
		fmt.Println("rebooting...")
		if err := c.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
		if err := c.Start(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
		fmt.Println("rebooted!")
	case "clean":
		fmt.Println("cleaning pid file...")
		if err := c.Wipe(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
		fmt.Println("cleaned!")
	default:
		fmt.Fprintf(os.Stderr, "Invalid argument: \"%s\"\n\n", os.Args[1])
		printHelp()
	}
}
