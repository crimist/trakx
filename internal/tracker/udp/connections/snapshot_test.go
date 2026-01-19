package connections

import (
	"bytes"
	"fmt"
	"net/netip"
	"testing"
)

func TestMarshalUnmarshal(t *testing.T) {
	connections := NewConnections(1, testTimeNever, testTimeNever)
	ipv4 := netip.MustParseAddrPort("1.1.1.1:1234")
	ipv6 := netip.MustParseAddrPort("[::1]:1234")

	id4, err := connections.Create(ipv4)
	if err != nil {
		t.Fatalf("failed to create ipv4 connection: %v", err)
	}
	id6, err := connections.Create(ipv6)
	if err != nil {
		t.Fatalf("failed to create ipv6 connection: %v", err)
	}

	data, err := connections.Marshal()
	if err != nil {
		t.Fatal("failed to marshal connections", err)
	}

	// Create new connections instance for restore
	restored := NewConnections(0, testTimeNever, testTimeNever)
	err = restored.Unmarshal(data)
	if err != nil {
		t.Fatal("failed to unmarshal connections", err)
	}

	if restored.associations[ipv4].ID != id4 {
		t.Errorf("ipv4 connection id = %v; want %v", restored.associations[ipv4].ID, id4)
	}
	if restored.associations[ipv6].ID != id6 {
		t.Errorf("ipv6 connection id = %v; want %v", restored.associations[ipv6].ID, id6)
	}
}

func TestSnapshotRestore(t *testing.T) {
	connections := NewConnections(10, testTimeNever, testTimeNever)
	ipv4 := netip.MustParseAddrPort("192.168.1.1:8080")
	ipv6 := netip.MustParseAddrPort("[2001:db8::1]:9090")

	id4, err := connections.Create(ipv4)
	if err != nil {
		t.Fatalf("failed to create ipv4 connection: %v", err)
	}
	id6, err := connections.Create(ipv6)
	if err != nil {
		t.Fatalf("failed to create ipv6 connection: %v", err)
	}

	var buf bytes.Buffer
	if err := connections.Snapshot(&buf); err != nil {
		t.Fatal("failed to create snapshot:", err)
	}

	restored := NewConnections(0, testTimeNever, testTimeNever)
	if err := restored.Restore(&buf); err != nil {
		t.Fatal("failed to restore snapshot:", err)
	}

	if len(restored.associations) != 2 {
		t.Errorf("restored connections count = %d; want 2", len(restored.associations))
	}

	if restored.associations[ipv4].ID != id4 {
		t.Errorf("ipv4 connection id = %v; want %v", restored.associations[ipv4].ID, id4)
	}
	if restored.associations[ipv6].ID != id6 {
		t.Errorf("ipv6 connection id = %v; want %v", restored.associations[ipv6].ID, id6)
	}

	// Verify timestamps are preserved
	if restored.associations[ipv4].TimeStamp != connections.associations[ipv4].TimeStamp {
		t.Errorf("ipv4 timestamp = %v; want %v", restored.associations[ipv4].TimeStamp, connections.associations[ipv4].TimeStamp)
	}
	if restored.associations[ipv6].TimeStamp != connections.associations[ipv6].TimeStamp {
		t.Errorf("ipv6 timestamp = %v; want %v", restored.associations[ipv6].TimeStamp, connections.associations[ipv6].TimeStamp)
	}
}

func TestEmptySnapshot(t *testing.T) {
	connections := NewConnections(0, testTimeNever, testTimeNever)

	var buf bytes.Buffer
	if err := connections.Snapshot(&buf); err != nil {
		t.Fatal("failed to create empty snapshot:", err)
	}

	restored := NewConnections(0, testTimeNever, testTimeNever)
	if err := restored.Restore(&buf); err != nil {
		t.Fatal("failed to restore empty snapshot:", err)
	}

	if len(restored.associations) != 0 {
		t.Errorf("restored connections count = %d; want 0", len(restored.associations))
	}
}

func TestLargeSnapshot(t *testing.T) {
	const numConnections = 1000
	connections := NewConnections(numConnections, testTimeNever, testTimeNever)

	// Create mixed IPv4 and IPv6 connections
	ids := make(map[netip.AddrPort]uint64, numConnections)
	for i := 0; i < numConnections; i++ {
		var addr netip.AddrPort
		if i%2 == 0 {
			// IPv4
			addr = netip.MustParseAddrPort(fmt.Sprintf("10.0.%d.%d:%d", i/256, i%256, 1024+i))
		} else {
			// IPv6
			addr = netip.MustParseAddrPort(fmt.Sprintf("[2001:db8::%x]:%d", i, 1024+i))
		}
		id, err := connections.Create(addr)
		if err != nil {
			t.Fatalf("failed to create connection %d: %v", i, err)
		}
		ids[addr] = id
	}

	var buf bytes.Buffer
	if err := connections.Snapshot(&buf); err != nil {
		t.Fatal("failed to create large snapshot:", err)
	}

	restored := NewConnections(0, testTimeNever, testTimeNever)
	if err := restored.Restore(&buf); err != nil {
		t.Fatal("failed to restore large snapshot:", err)
	}

	if len(restored.associations) != numConnections {
		t.Errorf("restored connections count = %d; want %d", len(restored.associations), numConnections)
	}

	// Verify all connections were restored correctly
	for addr, expectedID := range ids {
		entry, ok := restored.associations[addr]
		if !ok {
			t.Errorf("connection %v not found in restored snapshot", addr)
			continue
		}
		if entry.ID != expectedID {
			t.Errorf("connection %v id = %v; want %v", addr, entry.ID, expectedID)
		}
	}
}

func BenchmarkSnapshot(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			connections := NewConnections(size, testTimeNever, testTimeNever)

			// Create mixed IPv4 and IPv6 connections
			for i := 0; i < size; i++ {
				var addr netip.AddrPort
				if i%2 == 0 {
					addr = netip.MustParseAddrPort(fmt.Sprintf("10.0.%d.%d:%d", i/256, i%256, 1024+i))
				} else {
					addr = netip.MustParseAddrPort(fmt.Sprintf("[2001:db8::%x]:%d", i, 1024+i))
				}
				if _, err := connections.Create(addr); err != nil {
					b.Fatalf("failed to create connection: %v", err)
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				if err := connections.Snapshot(&buf); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkRestore(b *testing.B) {
	sizes := []int{100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			connections := NewConnections(size, testTimeNever, testTimeNever)

			// Create mixed IPv4 and IPv6 connections
			for i := 0; i < size; i++ {
				var addr netip.AddrPort
				if i%2 == 0 {
					addr = netip.MustParseAddrPort(fmt.Sprintf("10.0.%d.%d:%d", i/256, i%256, 1024+i))
				} else {
					addr = netip.MustParseAddrPort(fmt.Sprintf("[2001:db8::%x]:%d", i, 1024+i))
				}
				if _, err := connections.Create(addr); err != nil {
					b.Fatalf("failed to create connection: %v", err)
				}
			}

			var buf bytes.Buffer
			if err := connections.Snapshot(&buf); err != nil {
				b.Fatal(err)
			}
			data := buf.Bytes()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				restored := NewConnections(0, testTimeNever, testTimeNever)
				if err := restored.Restore(bytes.NewReader(data)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	connections := NewConnections(1000, testTimeNever, testTimeNever)

	// Create mixed IPv4 and IPv6 connections
	for i := 0; i < 1000; i++ {
		var addr netip.AddrPort
		if i%2 == 0 {
			addr = netip.MustParseAddrPort(fmt.Sprintf("10.0.%d.%d:%d", i/256, i%256, 1024+i))
		} else {
			addr = netip.MustParseAddrPort(fmt.Sprintf("[2001:db8::%x]:%d", i, 1024+i))
		}
		if _, err := connections.Create(addr); err != nil {
			b.Fatalf("failed to create connection: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := connections.Marshal(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	connections := NewConnections(1000, testTimeNever, testTimeNever)

	// Create mixed IPv4 and IPv6 connections
	for i := 0; i < 1000; i++ {
		var addr netip.AddrPort
		if i%2 == 0 {
			addr = netip.MustParseAddrPort(fmt.Sprintf("10.0.%d.%d:%d", i/256, i%256, 1024+i))
		} else {
			addr = netip.MustParseAddrPort(fmt.Sprintf("[2001:db8::%x]:%d", i, 1024+i))
		}
		if _, err := connections.Create(addr); err != nil {
			b.Fatalf("failed to create connection: %v", err)
		}
	}

	data, err := connections.Marshal()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		restored := NewConnections(0, testTimeNever, testTimeNever)
		if err := restored.Unmarshal(data); err != nil {
			b.Fatal(err)
		}
	}
}
