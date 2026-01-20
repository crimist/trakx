package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// UDP heartbeat packet for trakx.
var udpHeartbeatRequest = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0}
var udpHeartbeatOK = []byte{0xFF}

// Tracker manages the trakx tracker lifecycle.
type Tracker struct {
	bin          string
	config       string
	httpAddr     string
	udpAddr      string
	readyTimeout time.Duration

	mu         sync.RWMutex
	cachePath  string
	httpClient *http.Client
}

// NewTracker creates a new tracker manager.
func NewTracker(bin, config, httpAddr, udpAddr string, readyTimeout time.Duration) *Tracker {
	return &Tracker{
		bin:          bin,
		config:       config,
		httpAddr:     httpAddr,
		udpAddr:      udpAddr,
		readyTimeout: readyTimeout,
		httpClient:   &http.Client{Timeout: time.Second},
	}
}

// Start starts the tracker with the given goroutine configuration.
func (t *Tracker) Start(ctx context.Context, udpRoutines, httpRoutines int) error {
	env := os.Environ()
	env = append(env, fmt.Sprintf("TRAKX_UDP_ROUTINES=%d", udpRoutines))
	env = append(env, fmt.Sprintf("TRAKX_HTTP_ROUTINES=%d", httpRoutines))

	cmd := exec.CommandContext(ctx, t.bin, "--config", t.config, "start")
	cmd.Env = env
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("start tracker: %w", err)
	}

	if err := t.WaitReady(ctx); err != nil {
		_ = t.Stop(context.Background())
		return fmt.Errorf("tracker not ready: %w", err)
	}

	return nil
}

// Stop stops the tracker.
func (t *Tracker) Stop(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, t.bin, "--config", t.config, "stop")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Restart stops any running tracker, clears cache, and starts fresh.
// Returns an error if the tracker fails to start. Stop/clear errors are logged but not fatal.
func (t *Tracker) Restart(ctx context.Context, udpRoutines, httpRoutines int) error {
	if t.IsAlive() {
		if err := t.Stop(ctx); err != nil {
			slog.Warn("stop tracker", "error", err)
		}
		time.Sleep(time.Second)
	}

	if err := t.ClearCache(); err != nil {
		slog.Warn("clear cache", "error", err)
	}

	return t.Start(ctx, udpRoutines, httpRoutines)
}

// WaitReady waits for the tracker to become ready.
func (t *Tracker) WaitReady(ctx context.Context) error {
	deadline := time.Now().Add(t.readyTimeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if t.httpAddr != "" && t.checkHTTPReady() {
			return nil
		}
		if t.udpAddr != "" && t.checkUDPReady() {
			return nil
		}

		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for tracker readiness")
}

// IsAlive checks if the tracker is currently running.
func (t *Tracker) IsAlive() bool {
	if t.httpAddr != "" && t.checkHTTPReady() {
		return true
	}
	if t.udpAddr != "" && t.checkUDPReady() {
		return true
	}
	return false
}

func (t *Tracker) checkHTTPReady() bool {
	url := fmt.Sprintf("http://%s/heartbeat", t.httpAddr)
	resp, err := t.httpClient.Get(url)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func (t *Tracker) checkUDPReady() bool {
	conn, err := net.DialTimeout("udp", t.udpAddr, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Second))
	if _, err := conn.Write(udpHeartbeatRequest); err != nil {
		return false
	}

	buf := make([]byte, 16)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	return n == 1 && buf[0] == udpHeartbeatOK[0]
}

// ClearCache removes known trakx cache files without deleting the entire directory.
// This targets: pid, db, maximums/, and trakx-backup-*.sock files.
func (t *Tracker) ClearCache() error {
	cachePath := t.getCachePath()
	if cachePath == "" {
		cachePath = t.readCachePathFromConfig()
		if cachePath != "" {
			t.setCachePath(cachePath)
		}
	}
	if cachePath == "" {
		home, _ := os.UserHomeDir()
		if home != "" {
			cachePath = filepath.Join(home, ".cache", "trakx")
			t.setCachePath(cachePath)
		}
	}

	if cachePath == "" || cachePath == "/" || !filepath.IsAbs(cachePath) {
		return nil
	}

	var errs []error

	if err := os.Remove(filepath.Join(cachePath, "pid")); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove pid: %w", err))
	}

	if err := os.Remove(filepath.Join(cachePath, "db")); err != nil && !os.IsNotExist(err) {
		errs = append(errs, fmt.Errorf("remove db: %w", err))
	}

	if err := os.RemoveAll(filepath.Join(cachePath, "maximums")); err != nil {
		errs = append(errs, fmt.Errorf("remove maximums: %w", err))
	}

	sockFiles, err := filepath.Glob(filepath.Join(cachePath, "trakx-backup-*.sock"))
	if err != nil {
		errs = append(errs, fmt.Errorf("glob sock files: %w", err))
	}
	for _, f := range sockFiles {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("remove %s: %w", filepath.Base(f), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("clear cache: %v", errs)
	}
	return nil
}

func (t *Tracker) getCachePath() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cachePath
}

func (t *Tracker) setCachePath(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cachePath = path
}

type trakxConfig struct {
	Cache string `yaml:"cache"`
}

func (t *Tracker) readCachePathFromConfig() string {
	if t.config == "" {
		return ""
	}
	data, err := os.ReadFile(t.config)
	if err != nil {
		return ""
	}

	var cfg trakxConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ""
	}

	value := cfg.Cache
	if value == "" {
		return ""
	}
	if !filepath.IsAbs(value) {
		configDir := filepath.Dir(t.config)
		value = filepath.Join(configDir, value)
	}
	return value
}
