package stats

import "net/netip"

type noopCollector struct{}

func (s *noopCollector) Hit()                  {}
func (s *noopCollector) Connect()              {}
func (s *noopCollector) Announce()             {}
func (s *noopCollector) Scrape()               {}
func (s *noopCollector) AddSeeds(n int64)      {}
func (s *noopCollector) AddLeeches(n int64)    {}
func (s *noopCollector) ErrorResponse()        {}
func (s *noopCollector) IPs() IPCollector      { return &noopIPCollector{} }
func (s *noopCollector) GetSnapshot() Snapshot { return Snapshot{} }

type noopIPCollector struct{}

func (ipc *noopIPCollector) Inc(ip netip.Addr)               {}
func (ipc *noopIPCollector) Remove(ip netip.Addr)            {}
func (ipc *noopIPCollector) Replace(oldIP, newIP netip.Addr) {}
func (ipc *noopIPCollector) Total() int                      { return 0 }
