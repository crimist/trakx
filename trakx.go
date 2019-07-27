package main

import (
	"fmt"
	"os"

	"github.com/Syc0x00/Trakx/controller"
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
	case "run":
		fmt.Println("Running...")
		c.Run()
		fmt.Println("Ran!")
	case "start":
		fmt.Println("starting...")
		if err := c.Start(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
		fmt.Println("started!")
	case "stop":
		fmt.Println("stopping...")
		if err := c.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
		fmt.Println("stopped!")
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
		fmt.Println("reloading...")
		if err := c.Reload(); err != nil {
			fmt.Fprintf(os.Stderr, err.Error()+"\n")
			os.Exit(-1)
		}
		fmt.Println("reloaded")
	default:
		printHelp(os.Args[1])
	}
}
