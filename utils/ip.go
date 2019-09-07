package utils

import (
	"encoding/binary"
	"net"
)

// IPToInt converts ip net.IP to uint32
func IPToInt(ip net.IP) uint32 {
	if len(ip) == 16 { // IPv6
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

// IntToIP converts uint32 to a net.IP
func IntToIP(ipnum uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipnum)
	return ip
}
