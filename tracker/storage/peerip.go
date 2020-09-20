package storage

import (
	"strings"

	"github.com/pkg/errors"
)

type PeerIP [4]byte

// Set sets the ip from a string
func (ip *PeerIP) Set(s string) error {
	var digitpos, arpos = 0, 0
	var digitar [3]uint8

	if strings.Contains(s, ":") {
		return errors.New("ipv6 unsupported")
	}
	if len(s) > 15 {
		return errors.New("ip too long")
	}

	for pos, c := range s {
		// if were at end then record digit before parsing
		if pos == len(s)-1 {
			digitar[digitpos] = uint8(c - '0')
			digitpos++
		}

		if c == '.' || pos == len(s)-1 { // if we hit dot or are at the end of the string
			var n uint16 // because we check if it's over 15 we can guarantee they dont enter uint32 or bigger vals
			switch digitpos {
			case 3:
				n += uint16(digitar[0]*100) + uint16(digitar[1]*10) + uint16(digitar[2])
			case 2:
				n += uint16(digitar[0]*10) + uint16(digitar[1])
			case 1:
				n += uint16(digitar[0])
			default:
				return errors.New("no number before dot '.'")
			}

			if n > 255 {
				return errors.New("digit group > 255")
			}
			ip[arpos] = uint8(n)

			// move to next digits
			arpos++
			if arpos > 3 {
				return errors.New("more than 4 digit groups (more than three dots)")
			}

			// reset digits and move on
			digitpos = 0
			digitar = [3]uint8{0}
			continue
		}

		digitar[digitpos] = uint8(c - '0')
		digitpos++
		if digitpos > 3 {
			return errors.New("more than 3 digits before dot '.'")
		}
	}

	if ip[0] == 0 {
		return errors.New("invalid ip")
	}

	return nil
}
