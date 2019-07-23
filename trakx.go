package main

import (
	"flag"

	"github.com/Syc0x00/Trakx/tracker"
)

func main() {
	// Get flags
	prodFlag := flag.Bool("x", false, "Production mode")
	httpFlag := flag.Bool("http", true, "HTTP Tracker")
	udpFlag := flag.Bool("udp", true, "UDP Tracker")
	flag.Parse()

	tracker.Run(*prodFlag, *udpFlag, *httpFlag)
}
