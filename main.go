package main

import (
	"flag"
	"fmt"
	"runtime"
	"syscall"

	"github.com/Syc0x00/Trakx/tracker"
)

func main() {
	// Get flags
	prodFlag := flag.Bool("x", false, "Production mode")
	flag.Parse()

	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		panic(err)
	}

	if runtime.GOOS == "darwin" {
		limit.Cur = 24576
	} else {
		limit.Cur = limit.Max
	}

	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		panic(err)
	} else {
		fmt.Printf("Set limit to %v\n", limit.Cur)
	}

	tracker.Run(*prodFlag)
}
