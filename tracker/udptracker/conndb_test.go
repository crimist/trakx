package udptracker

import (
	"encoding/binary"
	"math/rand"
	"net/netip"
	"os"
	"testing"
	"time"

	"github.com/crimist/trakx/config"
)

const (
	connectionTimeout = 24 * time.Hour
)

func init() {
	// cache in local directory
	config.Config.SetLogLevel(config.ErrorLevel)
}

func TestConnectionDatabaseAdd(t *testing.T) {
	connDb := newConnectionDatabase(connectionTimeout)
	id4 := int64(1337)
	addrPort4 := netip.MustParseAddrPort("1.1.1.1:1234")
	id6 := int64(13337)
	addrPort6 := netip.MustParseAddrPort("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:1234")

	timestamp := time.Now().Unix()
	connDb.add(id4, addrPort4)
	connDb.add(id6, addrPort6)
	connInfo4 := connDb.connectionMap[addrPort4]
	connInfo6 := connDb.connectionMap[addrPort6]

	if connInfo4.TimeStamp != timestamp {
		t.Errorf("connInfo4.timeStamp = %v; want %v", connInfo4.TimeStamp, timestamp)
	}
	if connInfo4.ID != id4 {
		t.Errorf("connInfo4.id = %v; want %v", connInfo4.ID, id4)
	}
	if connInfo6.TimeStamp != timestamp {
		t.Errorf("connInfo6.timeStamp = %v; want %v", connInfo6.TimeStamp, timestamp)
	}
	if connInfo6.ID != id6 {
		t.Errorf("connInfo6.id = %v; want %v", connInfo6.ID, id6)
	}
}

func TestConnectionDatabaseCheck(t *testing.T) {
	connDb := newConnectionDatabase(connectionTimeout)
	id4 := int64(1337)
	addrPort4 := netip.MustParseAddrPort("1.1.1.1:1234")
	id6 := int64(13337)
	addrPort6 := netip.MustParseAddrPort("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:1234")

	connDb.add(id4, addrPort4)
	connDb.add(id6, addrPort6)

	if !connDb.check(id4, addrPort4) {
		t.Error("valid check() returned false; want true")
	}
	if !connDb.check(id6, addrPort6) {
		t.Error("valid check() returned false; want true")
	}
	if connDb.check(7331, addrPort4) {
		t.Error("invalid id check() returned true; want false")
	}
	if connDb.check(7331, addrPort6) {
		t.Error("invalid id check() returned true; want false")
	}
}

func TestConnectionDatabaseTrim(t *testing.T) {
	connDb := newConnectionDatabase(time.Nanosecond)
	id := int64(1337)
	addrPort := netip.MustParseAddrPort("1.1.1.1:1234")

	connDb.add(id, addrPort)
	time.Sleep(1 * time.Second)
	connDb.trim()

	if connDb.check(id, addrPort) {
		t.Error("valid check() returned true; want false")
	}
}

func TestConnectionDatabaseWriteLoad(t *testing.T) {
	filePath := "writetest.db"
	connDbWrite := newConnectionDatabase(connectionTimeout)
	connDbLoad := newConnectionDatabase(connectionTimeout)
	id4 := int64(1337)
	addrPort4 := netip.MustParseAddrPort("1.1.1.1:1234")
	id6 := int64(13337)
	addrPort6 := netip.MustParseAddrPort("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:1234")

	connDbWrite.add(id4, addrPort4)
	connDbWrite.add(id6, addrPort6)
	if err := connDbWrite.writeToFile(filePath); err != nil {
		t.Errorf("write() failed: %v", err)
	}
	if err := connDbLoad.loadFromFile(filePath); err != nil {
		t.Errorf("load() failed: %v", err)
	}

	if !connDbLoad.check(id4, addrPort4) {
		t.Error("valid check() after load() returned false; want true")
	}
	if !connDbLoad.check(id6, addrPort6) {
		t.Error("valid check() after load() returned false; want true")
	}

	if err := os.Remove(filePath); err != nil {
		t.Logf("failed to remove conndb file: %v", err)
	}
}

func newConnectionDatabaseWithCount(count int) *connectionDatabase {
	connDb := newConnectionDatabase(connectionTimeout)
	rand.Seed(time.Now().UnixNano())
	var buf [4]byte

	for n := 0; n < count; n++ {
		ip := rand.Uint32()
		binary.LittleEndian.PutUint32(buf[:], ip)
		addrPort4 := netip.AddrPortFrom(netip.AddrFrom4(buf), uint16(rand.Int31()))

		connDb.add(int64(n), addrPort4)
	}

	return connDb
}

func TestMarshallUnmarshallConnectionDatabase(t *testing.T) {
	connDB := newConnectionDatabase(10 * time.Minute)
	connDB.add(1234, netip.MustParseAddrPort("1.1.1.1:1234"))

	data, err := connDB.marshallBinary()
	if err != nil {
		t.Fatalf("marshallBinary threw err: %v", err)
	}

	connDB = newConnectionDatabase(10 * time.Minute)

	if err := connDB.unmarshallBinary(data); err != nil {
		t.Fatalf("unmarshallBinary threw err: %v", err)
	}
	if !connDB.check(1234, netip.MustParseAddrPort("1.1.1.1:1234")) {
		t.Fatalf("check failed, database = %v", connDB.connectionMap)
	}
}

func benchmarkConnectionDatabaseMarshallBinary(b *testing.B, count int) {
	connDb := newConnectionDatabaseWithCount(count)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := connDb.marshallBinary(); err != nil {
			b.Error("marshallBinary failed;", err)
		}
	}
}

func BenchmarkConnectionDatabaseMarshallBinary100(b *testing.B) {
	benchmarkConnectionDatabaseMarshallBinary(b, 100)
}
func BenchmarkConnectionDatabaseMarshallBinary1000(b *testing.B) {
	benchmarkConnectionDatabaseMarshallBinary(b, 1000)
}
func BenchmarkConnectionDatabaseMarshallBinary10000(b *testing.B) {
	benchmarkConnectionDatabaseMarshallBinary(b, 10000)
}

func benchmarkConnectionDatabaseUnmarshallBinary(b *testing.B, count int) {
	connDb := newConnectionDatabaseWithCount(count)
	data, err := connDb.marshallBinary()
	if err != nil {
		b.Error("marshallBinary failed;", err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := connDb.unmarshallBinary(data); err != nil {
			b.Error("unmarshallBinary failed;", err)
		}
	}
}

func BenchmarkConnectionDatabaseUnmarshallBinary100(b *testing.B) {
	benchmarkConnectionDatabaseUnmarshallBinary(b, 100)
}
func BenchmarkConnectionDatabaseUnmarshallBinary1000(b *testing.B) {
	benchmarkConnectionDatabaseUnmarshallBinary(b, 1000)
}
func BenchmarkConnectionDatabaseUnmarshallBinary10000(b *testing.B) {
	benchmarkConnectionDatabaseUnmarshallBinary(b, 10000)
}

func benchmarkConnectionDatabaseGobEncode(b *testing.B, count int) {
	connDb := newConnectionDatabaseWithCount(count)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if _, err := connDb.gobEncode(); err != nil {
			b.Error("gobEncode failed;", err)
		}
	}
}

func BenchmarkConnectionDatabaseGobEncode100(b *testing.B) {
	benchmarkConnectionDatabaseGobEncode(b, 100)
}
func BenchmarkConnectionDatabaseGobEncode1000(b *testing.B) {
	benchmarkConnectionDatabaseGobEncode(b, 1000)
}
func BenchmarkConnectionDatabaseGobEncode10000(b *testing.B) {
	benchmarkConnectionDatabaseGobEncode(b, 10000)
}

func benchmarkConnectionDatabaseGobDecode(b *testing.B, count int) {
	connDb := newConnectionDatabaseWithCount(count)
	data, err := connDb.gobEncode()
	if err != nil {
		b.Error("marshallBinary failed;", err)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := connDb.gobDecode(data); err != nil {
			b.Error("unmarshallBinary failed;", err)
		}
	}
}

func BenchmarkConnectionDatabaseGobDecode100(b *testing.B) {
	benchmarkConnectionDatabaseGobDecode(b, 100)
}
func BenchmarkConnectionDatabaseGobDecode1000(b *testing.B) {
	benchmarkConnectionDatabaseGobDecode(b, 1000)
}
func BenchmarkConnectionDatabaseGobDecode10000(b *testing.B) {
	benchmarkConnectionDatabaseGobDecode(b, 10000)
}
