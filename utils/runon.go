/*
Utils contains misc utility functions for Trakx.
*/
package utils

import (
	"time"
)

// RunOn will run the given function at the exact tick of the duration.
// For example, RunOn(1*time.Minute, f) would execute f on the minute, every minute.
func RunOn(duration time.Duration, run func()) {
	nextTick := time.Now().Truncate(duration)
	for {
		nextTick = nextTick.Add(duration)
		time.Sleep(time.Until(nextTick))
		run()
	}
}
