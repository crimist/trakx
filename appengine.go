// +build heroku

// Trakx runner
// For use with appengines

package main

import (
	"github.com/syc0x00/trakx/tracker"
)

func main() {
	tracker.Run()
}
