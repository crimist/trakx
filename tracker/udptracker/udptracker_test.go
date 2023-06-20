package udptracker

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/crimist/trakx/storage/inmemory"
	"github.com/crimist/trakx/tracker/udptracker/conncache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	testNetworkAddress = "127.0.0.1"
	testStartupDelay   = 10 * time.Millisecond
)

var testNetworkPort = 10000

func findOpenPort() int {
	for {
		udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", testNetworkPort))
		if err != nil {
			zap.L().Fatal("failed to resolve UDP address", zap.Int("port", testNetworkPort), zap.Error(err))
		}
		listener, err := net.ListenUDP("udp", udpAddr)
		if err != nil {
			zap.L().Debug("Port is already bound", zap.Int("port", testNetworkPort), zap.Error(err))
			testNetworkPort++
			continue
		}
		listener.Close()
		break
	}

	return testNetworkPort
}

func TestMain(m *testing.M) {
	loggerConfig := zap.NewDevelopmentConfig()
	logger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(loggerConfig.EncoderConfig), zapcore.Lock(os.Stdout), zap.NewAtomicLevelAt(zap.DebugLevel)))
	zap.ReplaceGlobals(logger)

	findOpenPort()

	peerDB, err := inmemory.NewInMemory(inmemory.Config{})
	if err != nil {
		zap.L().Fatal("UDP trakcer received shutdown")
	}
	connCache := conncache.NewConnectionCache(1, 1*time.Minute, 1*time.Minute, "")
	tracker := NewTracker(peerDB, connCache, nil, TrackerConfig{
		Validate:         true,
		DefaultNumwant:   10,
		MaximumNumwant:   100,
		Interval:         1,
		IntervalVariance: 0,
	})
	go func() {
		tracker.Serve(nil, testNetworkPort, 1)
		if err != nil {
			zap.L().Fatal("failed to serve tracker")
		}
	}()

	time.Sleep(testStartupDelay)
	m.Run()

	tracker.Shutdown()
}

func dialTestTracker() (*net.UDPConn, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", testNetworkAddress, testNetworkPort))
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve UDP address")
	}
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial UDP address")
	}
	return conn, nil
}

// const (
// 	testTimeout         = 500 * time.Millisecond
// 	announceUDPaddress  = "127.0.0.1:1337"
// 	announceUDPaddress6 = "[::1]:1337"
// )

// func TestUDPAnnounce(t *testing.T) {
// 	packet := make([]byte, 0xFFFF)
// 	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn, err := net.DialUDP("udp4", nil, addr)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn.SetWriteDeadline(time.Now().Add(testTimeout))
// 	conn.SetReadDeadline(time.Now().Add(testTimeout))

// 	c := udpprotocol.ConnectRequest{
// 		ProtcolID:     udpprotocol.ProtocolMagic,
// 		Action:        udpprotocol.ActionConnect,
// 		TransactionID: 1337,
// 	}

// 	data, err := c.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Read(packet); err != nil {
// 		t.Fatal(err)
// 	}

// 	cr := udpprotocol.ConnectResponse{}
// 	cr.Unmarshall(packet)

// 	if cr.Action == udpprotocol.ActionError {
// 		e := udpprotocol.Error{}
// 		if err := e.Unmarshall(packet); err != nil {
// 			t.Fatal("failed to unmarshall tracker error:", err)
// 		}
// 		t.Error("server error:", string(e.ErrorString))
// 	}

// 	if cr.TransactionID != c.TransactionID {
// 		t.Errorf("transactionID = %v, want %v", cr.TransactionID, c.TransactionID)
// 	}
// 	if cr.Action != udpprotocol.ActionConnect {
// 		t.Errorf("action = %v, want 0", cr.Action)
// 	}

// 	a := udpprotocol.Announce{
// 		ConnectionID:  cr.ConnectionID,
// 		Action:        udpprotocol.ActionAnnounce,
// 		TransactionID: 7331,
// 		InfoHash:      [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
// 		PeerID:        [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
// 		Downloaded:    100,
// 		Left:          100,
// 		Uploaded:      50,
// 		Event:         2,
// 		IP:            0,
// 		Key:           0xDEADBEEF,
// 		NumWant:       1,
// 		Port:          0xAABB,
// 	}

// 	data, err = a.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Read(packet); err != nil {
// 		t.Fatal(err)
// 	}

// 	ar := udpprotocol.AnnounceResp{}
// 	ar.Unmarshall(packet)

// 	if ar.Action == udpprotocol.ActionError {
// 		e := udpprotocol.Error{}
// 		if err := e.Unmarshall(packet); err != nil {
// 			t.Fatal("failed to unmarshall tracker error:", err)
// 		}
// 		t.Fatal("server error:", string(e.ErrorString))
// 	}

// 	if ar.TransactionID != a.TransactionID {
// 		t.Errorf("transactionID = %v, want %v", ar.TransactionID, a.TransactionID)
// 	}
// 	if ar.Action != udpprotocol.ActionAnnounce {
// 		t.Errorf("action = %v, want 1", ar.Action)
// 	}
// 	if ar.Leechers != 1 {
// 		t.Errorf("leechers = %v, want 1", ar.Leechers)
// 	}
// 	if ar.Seeders != 0 {
// 		t.Errorf("seeders = %v, want 1", ar.Seeders)
// 	}
// 	if len(ar.Peers) < 1 {
// 		t.Fatal("no peers in response")
// 	}
// 	if !bytes.Equal(ar.Peers[4:6], []byte{0xAA, 0xBB}) {
// 		t.Errorf("peer port = %#v; want {0xAA, 0xBB}", ar.Peers[4:6])
// 	}
// 	if !bytes.Equal(ar.Peers[0:4], []byte{127, 0, 0, 1}) {
// 		t.Errorf("peer ip = %v; want {127, 0, 0, 1}", ar.Peers[0:4])
// 	}
// }

// func TestUDPAnnounce6(t *testing.T) {
// 	packet := make([]byte, 0xFFFF)
// 	addr, err := net.ResolveUDPAddr("udp6", announceUDPaddress6)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn, err := net.DialUDP("udp6", nil, addr)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn.SetWriteDeadline(time.Now().Add(testTimeout))
// 	conn.SetReadDeadline(time.Now().Add(testTimeout))

// 	c := udpprotocol.Connect{
// 		ProtcolID:     udpprotocol.ProtocolMagic,
// 		Action:        udpprotocol.ActionConnect,
// 		TransactionID: 1337,
// 	}

// 	data, err := c.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Fatal(err)
// 	}
// 	_, err = conn.Read(packet)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	cr := udpprotocol.ConnectResp{}
// 	cr.Unmarshall(packet)

// 	if cr.Action == udpprotocol.ActionError {
// 		e := udpprotocol.Error{}
// 		e.Unmarshall(packet)
// 		t.Error("Tracker err:", string(e.ErrorString))
// 	}

// 	if cr.TransactionID != c.TransactionID {
// 		t.Error("Invalid transactionID should be", c.TransactionID, "but got", cr.TransactionID)
// 	}
// 	if cr.Action != udpprotocol.ActionConnect {
// 		t.Error("Invalid action should be 0 but got", cr.Action)
// 	}

// 	a := udpprotocol.Announce{
// 		ConnectionID:  cr.ConnectionID,
// 		Action:        udpprotocol.ActionAnnounce,
// 		TransactionID: 7331,
// 		InfoHash:      [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
// 		PeerID:        [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
// 		Downloaded:    100,
// 		Left:          100,
// 		Uploaded:      50,
// 		Event:         2,
// 		IP:            0,
// 		Key:           0xDEADBEEF,
// 		NumWant:       1,
// 		Port:          0xAABB,
// 	}

// 	data, err = a.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Error(err)
// 	}
// 	_, err = conn.Read(packet)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	ar := udpprotocol.AnnounceResp{}
// 	ar.Unmarshall(packet)

// 	if ar.Action == udpprotocol.ActionError {
// 		e := udpprotocol.Error{}
// 		e.Unmarshall(packet)
// 		t.Error("Tracker err:", string(e.ErrorString))
// 		return
// 	}

// 	if ar.TransactionID != a.TransactionID {
// 		t.Error("Invalid transactionID should be", a.TransactionID, "but got", ar.TransactionID)
// 	}
// 	if ar.Action != udpprotocol.ActionAnnounce {
// 		t.Error("Invalid action should be 1 but got", ar.Action)
// 	}
// 	if ar.Leechers != 1 {
// 		t.Error("Invalid leechers should be 1 but got", ar.Leechers)
// 	}
// 	if ar.Seeders != 0 {
// 		t.Error("Invalid seeders should be 1 but got", ar.Seeders)
// 	}

// 	if len(ar.Peers) < 1 {
// 		t.Error("No peers")
// 		return
// 	}

// 	if !bytes.Equal(ar.Peers[16:18], []byte{0xAA, 0xBB}) {
// 		t.Errorf("peer port = %#v, want {0xAA, 0xBB}", ar.Peers[16:18])
// 	}
// 	if !bytes.Equal(ar.Peers[0:16], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}) {
// 		t.Errorf("peer ip = %v; want {0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}", ar.Peers[0:16])
// 	}
// }

// func TestUDPBadAction(t *testing.T) {
// 	packet := make([]byte, 0xFFFF)
// 	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn, err := net.DialUDP("udp", nil, addr)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn.SetWriteDeadline(time.Now().Add(testTimeout))
// 	conn.SetReadDeadline(time.Now().Add(testTimeout))

// 	c := udpprotocol.Connect{
// 		ProtcolID:     udpprotocol.ProtocolMagic,
// 		Action:        udpprotocol.ActionConnect,
// 		TransactionID: 1337,
// 	}

// 	data, err := c.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Fatal(err)
// 	}
// 	_, err = conn.Read(packet)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	cr := udpprotocol.ConnectResp{}
// 	cr.Unmarshall(packet)

// 	c = udpprotocol.Connect{
// 		ProtcolID:     cr.ConnectionID,
// 		Action:        0xBAD,
// 		TransactionID: 0xDEAD,
// 	}

// 	data, err = c.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Error(err)
// 	}
// 	s, err := conn.Read(packet)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	e := udpprotocol.Error{}
// 	e.Unmarshall(packet[:s])

// 	if !bytes.Equal(e.ErrorString, []byte("bad action")) {
// 		t.Error("Tracker err should be 'bad action' but got:", string(e.ErrorString))
// 	}
// }

// func TestUDPBadConnID(t *testing.T) {
// 	packet := make([]byte, 0xFFFF)
// 	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn, err := net.DialUDP("udp", nil, addr)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn.SetWriteDeadline(time.Now().Add(testTimeout))
// 	conn.SetReadDeadline(time.Now().Add(testTimeout))

// 	a := udpprotocol.Announce{
// 		ConnectionID:  0xBAD, // bad connid
// 		Action:        udpprotocol.ActionAnnounce,
// 		TransactionID: 0xDEAD,
// 	}

// 	data, err := a.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Fatal(err)
// 	}
// 	s, err := conn.Read(packet)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	e := udpprotocol.Error{}
// 	e.Unmarshall(packet[:s])

// 	if !bytes.Equal(e.ErrorString, []byte("bad connid")) {
// 		t.Error("Tracker err should be 'bad connid' but got:", string(e.ErrorString))
// 	}
// }

// func TestUDPBadPort(t *testing.T) {
// 	packet := make([]byte, 0xFFFF)
// 	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn, err := net.DialUDP("udp", nil, addr)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn.SetWriteDeadline(time.Now().Add(testTimeout))
// 	conn.SetReadDeadline(time.Now().Add(testTimeout))

// 	c := udpprotocol.Connect{
// 		ProtcolID:     udpprotocol.ProtocolMagic,
// 		Action:        udpprotocol.ActionConnect,
// 		TransactionID: 1337,
// 	}

// 	data, err := c.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Fatal(err)
// 	}
// 	_, err = conn.Read(packet)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	cr := udpprotocol.ConnectResp{}
// 	cr.Unmarshall(packet)

// 	a := udpprotocol.Announce{
// 		ConnectionID:  cr.ConnectionID,
// 		Action:        udpprotocol.ActionAnnounce,
// 		TransactionID: 7331,
// 		InfoHash:      [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
// 		PeerID:        [20]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14},
// 		Downloaded:    100,
// 		Left:          100,
// 		Uploaded:      50,
// 		Event:         2,
// 		IP:            0,
// 		Key:           0xDEADBEEF,
// 		NumWant:       1,
// 		Port:          0,
// 	}

// 	data, err = a.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if _, err = conn.Write(data); err != nil {
// 		t.Fatal(err)
// 	}
// 	s, err := conn.Read(packet)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	e := udpprotocol.Error{}
// 	e.Unmarshall(packet[:s])

// 	if !bytes.Equal(e.ErrorString, []byte("bad port")) {
// 		t.Error("Tracker err should be 'bad port' but got:", string(e.ErrorString))
// 	}
// }

// func TestUDPTransactionID(t *testing.T) {
// 	packet := make([]byte, 0xFF)
// 	addr, err := net.ResolveUDPAddr("udp4", announceUDPaddress)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn, err := net.DialUDP("udp", nil, addr)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	conn.SetWriteDeadline(time.Now().Add(testTimeout))
// 	conn.SetReadDeadline(time.Now().Add(testTimeout))

// 	c := udpprotocol.Connect{
// 		ProtcolID:     udpprotocol.ProtocolMagic,
// 		Action:        udpprotocol.ActionConnect,
// 		TransactionID: 0xBAD,
// 	}
// 	data, err := c.Marshall()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	for i := 0; i < 1000; i++ {
// 		if _, err = conn.Write(data); err != nil {
// 			t.Fatal(err)
// 		}
// 		size, err := conn.Read(packet)
// 		if err != nil {
// 			t.Fatal(err)
// 		}

// 		if size != 16 {
// 			e := udpprotocol.Error{}
// 			e.Unmarshall(packet)
// 			t.Error(i, "Tracker err:", string(e.ErrorString))
// 		}

// 		cr := udpprotocol.ConnectResp{}
// 		cr.Unmarshall(packet)

// 		if cr.TransactionID != 0xBAD {
// 			t.Error(i, "Tracker err: tid should be", 0xBAD, "but got", cr.TransactionID)
// 		}
// 	}
// }
