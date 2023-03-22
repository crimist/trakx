//go:build heroku
// +build heroku

// This is the entrypoint for app engines such as Heroku, GAE, etc.
// You can modify the build tags above to suit the app engine if your choice

// This entrypoint doesn't do any of the CLI daemon behavior of `cli_entry.go` and instead immediatly executes the tracker
// You can customize the trackers configuration on your app engine by setting the appropriate environment variables

package main

import (
	"github.com/crimist/trakx/tracker"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	conf, err := config.Load()
	if err != nil {
		zap.L().Fatal("failed to load configuration", zap.Error(err))
	}

	tracker.Run(conf)
}
