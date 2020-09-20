package storage

import (
	"testing"

	"github.com/pkg/errors"
)

func TestPeerIPSet(t *testing.T) {
	var cases = []struct {
		input  string
		output PeerIP
		err    error
	}{
		{"1.1.1.1", PeerIP{1, 1, 1, 1}, nil},
		{"255.255.255.255", PeerIP{255, 255, 255, 255}, nil},
		{"1.255.1.255", PeerIP{1, 255, 1, 255}, nil},
		{"256.0.0.0", PeerIP{0, 0, 0, 0}, errors.New("digit group > 255")},
		{"0.1.1.1", PeerIP{0, 0, 0, 0}, errors.New("invalid ip")},
		{"1.1.1.1.1", PeerIP{0, 0, 0, 0}, errors.New("more than 4 digit groups (more than three dots)")},
		{"1234567890123456", PeerIP{0, 0, 0, 0}, errors.New("ip too long")},
		{".1.1.1", PeerIP{0, 0, 0, 0}, errors.New("no number before dot '.'")},
		{"1.1..1.1", PeerIP{0, 0, 0, 0}, errors.New("no number before dot '.'")},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			var ip PeerIP
			err := ip.Set(c.input)
			if err != nil {
				// if we wanted to error and they don't match
				if c.err != nil && err.Error() != c.err.Error() {
					t.Errorf("failed, error mismatch / unexpected error: '%v' / '%v'", err, c.err)
				}
			} else {
				if ip != c.output {
					t.Error("failed, output mismatch", ip, c.output)
				}
			}
		})
	}
}
