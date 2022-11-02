//go:build !fast

package stats

import (
	"net/netip"
)

const fast = false

func (ipstats *ipStats) Total() int {
	return len(ipstats.submap)
}

func (ipstats *ipStats) Delete(ip netip.Addr) {
	delete(ipstats.submap, ip)
}

func (ipstats *ipStats) Inc(ip netip.Addr) {
	ipstats.submap[ip]++
}

func (ipstats *ipStats) Dec(ip netip.Addr) {
	ipstats.submap[ip]--
}

// Remove decrements the IP and removes it if it's 0
func (ipstats *ipStats) Remove(ip netip.Addr) {
	ipstats.submap[ip]--
	if ipstats.submap[ip] == 0 {
		delete(ipstats.submap, ip)
	}
}
