// +build heroku

package http

import "bytes"

func getForwarded(data []byte) (bool, []byte) {
	s := bytes.Index(data, []byte("X-Forwarded-For"))
	if s == -1 {
		return false, nil
	}
	e := bytes.Index(data[s:], []byte("\r\n"))
	if e == -1 {
		return true, nil
	}

	// Check that there's something in the field
	if ((s + e) - (s + 17)) < 7 {
		return true, nil
	}

	ips := data[s+17 : s+e]

	// Check for multi and get farther right if so
	if bytes.Contains(ips, []byte(",")) {
		comma := bytes.LastIndex(ips, []byte(","))
		return true, ips[comma+2:]
	}

	return true, ips
}
