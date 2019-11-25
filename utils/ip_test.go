package utils

import (
	"bytes"
	"net"
	"testing"
)

func TestIPToInt(t *testing.T) {
	local := IPToInt(net.IPv4(127, 0, 0, 1))
	if local != 2130706433 {
		t.Error("127.0.0.1 failed")
	}
}

func TestIntToIP(t *testing.T) {
	t.Skip("Broken")

	local := IntToIP(2130706433) // 127.0.0.1
	if bytes.Compare(local, net.IPv4(127, 0, 0, 1)) != 0 {
		t.Error("127.0.0.1 failed")
	}
}
