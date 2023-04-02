//go:build !nostats

package stats

import (
	"net/netip"
)

func (ipstats *ipStatistics) IPs() int {
	return len(ipstats.ips)
}

func (ipstats *ipStatistics) Delete(ip netip.Addr) {
	delete(ipstats.ips, ip)
}

func (ipstats *ipStatistics) Inc(ip netip.Addr) {
	ipstats.ips[ip]++
}

func (ipstats *ipStatistics) Dec(ip netip.Addr) {
	ipstats.ips[ip]--
}

// Remove decrements the IP and removes it if it's 0
func (ipstats *ipStatistics) Remove(ip netip.Addr) {
	ipstats.ips[ip]--
	if ipstats.ips[ip] == 0 {
		delete(ipstats.ips, ip)
	}
}
