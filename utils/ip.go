package utils

import (
	"encoding/binary"
	"net"
)

// IPToInt converts ip string to uint32 address
func IPToInt(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}
