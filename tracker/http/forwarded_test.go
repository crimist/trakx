// +build !heroku

package http

import "testing"

func TestForwarded(t *testing.T) {
	t.Log("Skipping forwarded tests - appengine disabled")
}
