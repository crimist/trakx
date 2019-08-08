// +build !heroku

// Trakx controller
// For use on a server
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/syc0x00/trakx/controller"
)

const (
	trakxRoot  = "~/.trakx/"
	trakxPerms = 0740
)

func printHelp(arg string) {
	if arg != "" {
		fmt.Fprintf(os.Stderr, "Invalid argument: \"%s\"\n\n", arg)
	}
	help := "Trakx commands:\n"
	help += fmt.Sprintf("  %-12s Checks if Trakx is running\n", "status")
	help += fmt.Sprintf("  %-12s Runs Trakx if it closes\n", "watcher")
	help += fmt.Sprintf("  %-12s Runs Trakx (doesn't return)\n", "run")
	help += fmt.Sprintf("  %-12s Starts Trakx as a service\n", "start")
	help += fmt.Sprintf("  %-12s Stops Trakx service\n", "stop")
	help += fmt.Sprintf("  %-12s Restarts Trakx service\n", "restart")
	help += fmt.Sprintf("  %-12s Wipes trakx pid file\n", "wipe")
	help += fmt.Sprintf("  %-12s Reloads the Trakx config\n", "reload")
	help += "Usage:\n"
	help += fmt.Sprintf("  %s <command>\n", os.Args[0])
	help += "Example:\n"
	help += fmt.Sprintf("  %s run\n", os.Args[0])

	fmt.Print(help)
}

func main() {
	if len(os.Args) < 2 {
		printHelp("")
		return
	}

	c, err := controller.NewController(trakxRoot, trakxPerms)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(-1)
	}

	switch os.Args[1] {
	case "status":
		if c.IsRunning() {
			fmt.Println("Trakx is running")
		} else {
			fmt.Println("Trakx is not running")
		}
	case "watcher":
		for {
			if !c.IsRunning() {
				if err := c.Start(); err != nil {
					fmt.Fprintf(os.Stderr, err.Error()+"\n")
					os.Exit(-1)
				}
				// Wait 20 seconds to let it set up
				time.Sleep(20 * time.Second)
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
	case "wipe":
		fmt.Println("wiping...")
		if err := c.Wipe(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
		fmt.Println("wiped...")
	case "reload":
		if err := c.Reload(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
	default:
		printHelp(os.Args[1])
	}
}
