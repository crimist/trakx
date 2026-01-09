package udpprotocol

import (
	"strconv"
	"testing"
)

func TestActionValid(t *testing.T) {
	var cases = []struct {
		action   Action
		expected bool
	}{
		{-1, false},
		{ActionConnect, true},
		{ActionAnnounce, true},
		{ActionScrape, true},
		{ActionError, true},
		{ActionHeartbeat, true},
		{5, false},
	}

	for _, c := range cases {
		t.Run(strconv.Itoa(int(c.action)), func(t *testing.T) {
			result := c.action.Valid()
			if result != c.expected {
				t.Errorf("action %v valid = %v; want %v", c.action, result, c.expected)
			}
		})
	}
}
