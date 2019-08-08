// +build heroku

// Trakx runner
// For use on an app engine
package main

import (
	"github.com/syc0x00/trakx/tracker"
)

func main() {
	tracker.Run()
}
