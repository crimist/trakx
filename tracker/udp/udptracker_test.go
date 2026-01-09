package udp

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage"
	"github.com/crimist/trakx/storage/database"
	"github.com/crimist/trakx/tracker"
	"github.com/crimist/trakx/tracker/udp/connections"
	"github.com/crimist/trakx/tracker/udp/udpprotocol"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	testNetAddress4      = "127.0.0.1"
	testNetAddress6      = "::1"
	testMockStartupDelay = 10 * time.Millisecond
)

var (
	testTrackerConfig = tracker.TrackerConfig{
		DefaultNumwant:   2,
		MaximumNumwant:   3,
		Interval:         1,
		IntervalVariance: 0,
	}
	testNetworkPort = 10000
	noopCollector   = stats.NewCollectors(false, false, 0)
)

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
	zap.L().Debug("Found open port for UDP tracker", zap.Int("port", testNetworkPort))

	peerDB, err := database.NewDatabase(database.Config{
		Collector:          noopCollector,
		PeerlistMaxNumwant: testTrackerConfig.MaximumNumwant,
	})
	if err != nil {
		zap.L().Fatal("UDP tracker received shutdown")
	}
	connections := connections.NewConnections(1, 1*time.Minute, 1*time.Minute)
	tracker := NewTracker(peerDB, testTrackerConfig, noopCollector, connections, true)
	go func() {
		err = tracker.Serve(nil, testNetworkPort, 1)
		if err != nil {
			zap.L().Fatal("failed to serve tracker")
		}
	}()

	time.Sleep(testMockStartupDelay)
	m.Run()

	tracker.Shutdown()
}

func dialMockTracker(address string) (*net.UDPConn, error) {
	addr := &net.UDPAddr{
		IP:   net.ParseIP(address),
		Port: testNetworkPort,
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial UDP address")
	}
	return conn, nil
}

func TestUnregisteredConnection(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	errorResp := announceError(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  0xBAD,
		Action:        udpprotocol.ActionAnnounce,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          4096,
	})

	if !bytes.Equal(errorResp.ErrorString, []byte(fatalUnregisteredConnection)) {
		t.Errorf("Expected error = %s; got %s", string(fatalUnregisteredConnection), string(errorResp.ErrorString))
	}
}

func TestBadAction(t *testing.T) {
	conn, err := dialMockTracker(testNetAddress4)
	if err != nil {
		t.Fatal("failed to dial mock tracker", err)
	}

	connectResp := connect(t, conn, udpprotocol.ConnectRequest{
		ProtocolID:    udpprotocol.ProtocolMagic,
		Action:        udpprotocol.ActionConnect,
		TransactionID: 1,
	})

	errorResp := announceError(t, conn, udpprotocol.AnnounceRequest{
		ConnectionID:  connectResp.ConnectionID,
		Action:        5,
		TransactionID: 1,
		InfoHash:      storage.Hash{},
		PeerID:        storage.PeerID{},
		Downloaded:    1000,
		Left:          1000,
		Uploaded:      1000,
		Event:         udpprotocol.EventStarted,
		IP:            0,
		Key:           0x1337,
		NumWant:       50,
		Port:          4096,
	})

	if !bytes.Equal(errorResp.ErrorString, []byte(fatalInvalidAction)) {
		t.Errorf("Expected error = %v; got %v", fatalInvalidAction, errorResp.ErrorString)
	}
}
