package conncache

import (
	"fmt"
	"net/netip"
	"testing"
	"time"
)

func TestSet(t *testing.T) {
	var cases = []struct {
		name     string
		id       int64
		addrPort netip.AddrPort
	}{
		{"ipv4", 1, netip.MustParseAddrPort("1.1.1.1:1234")},
		{"ipv6", 2, netip.MustParseAddrPort("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:1234")},
	}
	cache := NewConnectionCache(len(cases), 1*time.Minute, 1*time.Minute, "")
	nowUnix := time.Now().Unix()

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			cache.Set(testCase.id, testCase.addrPort)

			entry := cache.entries[testCase.addrPort]

			if entry.TimeStamp != nowUnix {
				t.Errorf("entry timestamp = %v; want %v", entry.TimeStamp, nowUnix)
			}
			if entry.ID != testCase.id {
				t.Errorf("entry id = %v; want %v", entry.ID, testCase.id)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	cache := NewConnectionCache(1, 1*time.Minute, 1*time.Minute, "")
	addrPort := netip.MustParseAddrPort("1.1.1.1:1234")
	cache.Set(1, addrPort)

	if !cache.Validate(1, addrPort) {
		t.Error("cache validate returned false; want true")
	}
}

func TestEntryCount(t *testing.T) {
	const entries = 10
	cache := NewConnectionCache(entries, 1*time.Minute, 1*time.Minute, "")

	for i := int64(0); i < entries; i++ {
		cache.Set(i, netip.MustParseAddrPort(fmt.Sprintf("1.1.1.%d:1234", i)))

		if int64(cache.EntryCount()) != i+1 {
			t.Errorf("cache entry count = %v; want %v", cache.EntryCount(), i+1)
		}
	}
}

func TestEvictExpiredEntries(t *testing.T) {
	cacheExpireShort := NewConnectionCache(1, 1*time.Microsecond, 1*time.Minute, "")
	cacheExpireLong := NewConnectionCache(1, 1*time.Minute, 1*time.Minute, "")

	addrPort := netip.MustParseAddrPort("1.1.1.1:1234")
	cacheExpireShort.Set(1, addrPort)
	cacheExpireLong.Set(1, addrPort)
	time.Sleep(1 * time.Second)
	cacheExpireShort.evictExpiredEntries()
	cacheExpireLong.evictExpiredEntries()

	if _, ok := cacheExpireShort.entries[addrPort]; ok {
		t.Error("cache entry did not expire when it should have")
	}
	if _, ok := cacheExpireLong.entries[addrPort]; !ok {
		t.Error("cache entry expired when it it should not have")
	}
}
