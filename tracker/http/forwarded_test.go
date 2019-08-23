// +build !heroku

package http

import "testing"

func TestForwarded(t *testing.T) {
	t.Log("Appengine disabled")
}
