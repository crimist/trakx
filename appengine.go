// +build heroku

// Trakx runner
// For use with appengines

package main

import (
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/syc0x00/trakx/tracker"
)

func main() {
	tracker.Run()
}
