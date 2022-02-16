//go:build heroku
// +build heroku

// Trakx tracker run entrypoint for heroku (or any app engine)

package main

import (
	"github.com/crimist/trakx/tracker"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	tracker.Run()
}
