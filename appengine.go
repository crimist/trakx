// +build heroku

// Trakx runner
// For use with appengines

package main

import (
	_ "github.com/heroku/x/hmetrics/onload"
	"github.com/crimist/trakx/tracker"
)

func main() {
	tracker.Run()
}
