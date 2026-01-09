package http

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/storage/database"
	"github.com/crimist/trakx/tracker"
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
		Interval:         10,
		IntervalVariance: 0,
	}
	testNetworkPort = 10000
)

func findOpenPort() int {
	for {
		tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", testNetworkPort))
		if err != nil {
			zap.L().Fatal("failed to resolve TCP address", zap.Int("port", testNetworkPort), zap.Error(err))
		}
		listener, err := net.ListenTCP("tcp", tcpAddr)
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
	zap.L().Debug("Found open port for HTTP tracker", zap.Int("port", testNetworkPort))

	pools.Initialize(int(testTrackerConfig.MaximumNumwant))

	peerDB, err := database.NewDatabase(database.Config{
		Collector: stats.NewCollectors(false, false, 0),
	})
	if err != nil {
		zap.L().Fatal("TCP tracker received shutdown")
	}

	servePath, err := filepath.Abs(".")
	if err != nil {
		zap.L().Fatal("failed to get absolute path", zap.Error(err))
	}

	collector := stats.NewCollectors(false, false, 0)
	tracker := NewTracker(peerDB, testTrackerConfig, collector, servePath, 100*time.Millisecond, 100*time.Millisecond)
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

func dialMockTracker(address string) (*net.TCPConn, error) {
	addr := &net.TCPAddr{
		IP:   net.ParseIP(address),
		Port: testNetworkPort,
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial TCP address")
	}
	return conn, nil
}

func writeMockTracker(address string, msg string) (string, error) {
	conn, err := dialMockTracker(address)
	if err != nil {
		return "", errors.Wrap(err, "failed to dial mock tracker")
	}
	defer conn.Close()

	_, err = conn.Write([]byte(msg))
	if err != nil {
		return "", errors.Wrap(err, "failed to write to connection")
	}

	buf := make([]byte, 65535)
	n, err := conn.Read(buf)
	if err != nil {
		return "", errors.Wrap(err, "failed to read from connection")
	}
	buf = buf[:n]

	return string(buf), nil
}

func TestInvalidHTTP(t *testing.T) {
	resp, err := writeMockTracker(testNetAddress4, "Not a valid http request")
	if err != nil {
		t.Fatal(err)
	}

	if resp != "HTTP/1.1 400\r\n\r\n" {
		t.Errorf("Expected code 400, got %s", resp)
	}
}

func TestNonGET(t *testing.T) {
	resp, err := http.Post(fmt.Sprintf("http://%s:%d", testNetAddress4, testNetworkPort), "", nil)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected code 400, got %d", resp.StatusCode)
	}
}

func TestInvalidPath(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/invalid", testNetAddress4, testNetworkPort))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected code 404, got %d", resp.StatusCode)
	}
}

func TestHeartbeat(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/heartbeat", testNetAddress4, testNetworkPort))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected code 200, got %d", resp.StatusCode)
	}
}

func TestServe(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/.", testNetAddress4, testNetworkPort))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 403 {
		t.Errorf("Expected code 403, got %d", resp.StatusCode)
	}

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/../", testNetAddress4, testNetworkPort))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 403 {
		t.Errorf("Expected code 403, got %d", resp.StatusCode)
	}

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/../../../../../../../../../../../etc/passwd", testNetAddress4, testNetworkPort))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected code 404, got %d", resp.StatusCode)
	}

	resp, err = http.Get(fmt.Sprintf("http://%s:%d/httptracker_test.go", testNetAddress4, testNetworkPort))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected code 200, got %d", resp.StatusCode)
	}

	body := make([]byte, 65535)
	n, err := resp.Body.Read(body)
	if err != nil {
		t.Fatal("Failed to read response body", err)
	}
	body = body[:n]

	if string(body[:13]) != "package http\n" {
		t.Errorf("Expected file to read 'package http\\n', got %s", body)
	}
}
