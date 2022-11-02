//go:build fast

package stats

const fast = true

func (ipstats *IPStats) Len() int         { return -1 }
func (ipstats *IPStats) Delete(ip PeerIP) {}
func (ipstats *IPStats) Inc(ip PeerIP)    {}
func (ipstats *IPStats) Dec(ip PeerIP)    {}
func (ipstats *IPStats) Remove(ip PeerIP) {}
