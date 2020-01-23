// +build heroku
// Trakx direct run for use with appengines

package main

import (
	"github.com/crimist/trakx/tracker"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	tracker.Run()
}
