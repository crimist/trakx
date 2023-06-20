//go:build nostats

package stats

import (
	"net/netip"
)

const fast = true

func (ipstats *ipStats) Total() int           { return -1 }
func (ipstats *ipStats) Delete(ip netip.Addr) {}
func (ipstats *ipStats) Inc(ip netip.Addr)    {}
func (ipstats *ipStats) Dec(ip netip.Addr)    {}
func (ipstats *ipStats) Remove(ip netip.Addr) {}
