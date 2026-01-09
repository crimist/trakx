package connections

import (
	"fmt"
	"net/netip"
	"testing"
	"time"
)

const testTimeNever = 1 * time.Minute
const testTimeInstant = 1 * time.Microsecond

func TestCreate(t *testing.T) {
	var cases = []struct {
		name     string
		addrPort netip.AddrPort
	}{
		{"ipv4", netip.MustParseAddrPort("1.1.1.1:1234")},
		{"ipv6", netip.MustParseAddrPort("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:1234")},
	}
	connections := NewConnections(len(cases), testTimeNever, testTimeNever)

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			epoch := time.Now().Unix()
			id := connections.Create(testCase.addrPort)

			entry := connections.associations[testCase.addrPort]
			if entry.TimeStamp != epoch {
				t.Errorf("entry timestamp = %v; want %v", entry.TimeStamp, epoch)
			}
			if entry.ID != id {
				t.Errorf("entry id = %v; want %v", entry.ID, id)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	connections := NewConnections(1, testTimeNever, testTimeNever)
	addrPort := netip.MustParseAddrPort("1.1.1.1:1234")
	id := connections.Create(addrPort)

	if !connections.Validate(addrPort, id) {
		t.Error("cache validate returned false; want true")
	}
}

func TestEntries(t *testing.T) {
	const entries = 10
	connections := NewConnections(entries, testTimeNever, testTimeNever)

	for i := int64(0); i < entries; i++ {
		connections.Create(netip.MustParseAddrPort(fmt.Sprintf("1.1.1.%d:1234", i)))

		if int64(connections.Entries()) != i+1 {
			t.Errorf("cache entry count = %v; want %v", connections.Entries(), i+1)
		}
	}
}

// sleep 1 second since were using second accuracy for epoch

func TestGarbageCollector(t *testing.T) {
	connections := NewConnections(1, testTimeInstant, testTimeNever)
	addrPort := netip.MustParseAddrPort("1.1.1.1:1234")

	connections.Create(addrPort)
	time.Sleep(1 * time.Second)
	connections.garbageCollector()

	if _, ok := connections.associations[addrPort]; ok {
		t.Error("connection entry exists when it should have expired")
	}
}

func TestGarbageCollectorCron(t *testing.T) {
	connections := NewConnections(1, testTimeInstant, 100*time.Millisecond)
	addrPort := netip.MustParseAddrPort("1.1.1.1:1234")

	connections.Create(addrPort)
	time.Sleep(1 * time.Second)

	if _, ok := connections.associations[addrPort]; ok {
		t.Error("connection entry exists when it should have expired")
	}
}
