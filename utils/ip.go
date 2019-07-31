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

// IntToIP converts int to ip bytes
func IntToIP(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}
