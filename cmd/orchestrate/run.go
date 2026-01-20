package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Timeout constants for protocol operations.
const (
	readyTimeout   = 60 * time.Second
	stoppedTimeout = 30 * time.Second
)

// Executor handles benchmark execution for both local and distributed modes.
type Executor struct {
	tracker  *Tracker // nil for client mode
	conn     *Conn    // nil for local mode
	benchBin string
	target   string
	statsURL string
	output   string
}

// NewLocalExecutor creates an executor for local mode.
func NewLocalExecutor(tracker *Tracker, benchBin, target, output string) *Executor {
	return &Executor{
		tracker:  tracker,
		benchBin: benchBin,
		target:   target,
		output:   output,
	}
}

// NewClientExecutor creates an executor for distributed client mode.
func NewClientExecutor(conn *Conn, benchBin, target, statsURL, output string) *Executor {
	return &Executor{
		conn:     conn,
		benchBin: benchBin,
		target:   target,
		statsURL: statsURL,
		output:   output,
	}
}

// Run executes a single benchmark run.
func (e *Executor) Run(ctx context.Context, run BenchmarkRun) RunResult {
	start := time.Now()
	result := RunResult{
		RunID:       run.Name,
		Mode:        run.Mode,
		Routines:    run.Routines,
		OtherRoutes: run.OtherRoutines,
	}

	udpRoutines, httpRoutines := DetermineRoutines(run)

	if err := e.startTracker(ctx, udpRoutines, httpRoutines, run); err != nil {
		return e.fail(start, &result, "start tracker: %v", err)
	}
	defer e.stopTracker(ctx)

	slog.Info("  Running benchmark...")

	resultFile := filepath.Join(e.output, fmt.Sprintf("%s.json", run.Name))
	result.ResultFile = resultFile

	benchArgs := BuildBenchArgs(run, e.target, resultFile, e.statsURL)
	cmd := exec.CommandContext(ctx, e.benchBin, benchArgs...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	benchErr := cmd.Run()

	// For client mode, always notify server before checking bench result.
	if e.conn != nil {
		errorMsg := ""
		if benchErr != nil {
			errorMsg = benchErr.Error()
		}
		if err := e.notifyDone(udpRoutines, httpRoutines, resultFile, benchErr == nil, errorMsg); err != nil {
			return e.fail(start, &result, "notify done: %v", err)
		}
	}

	if benchErr != nil {
		return e.fail(start, &result, "benchmark failed: %v", benchErr)
	}

	summary, raw, err := LoadResultFile(resultFile)
	if err != nil {
		if os.IsNotExist(err) {
			result.Status = "incomplete"
			result.Error = "result file not created (benchmark may have crashed early)"
			result.DurationMS = time.Since(start).Milliseconds()
			return result
		}
		return e.fail(start, &result, "load results: %v", err)
	}

	result.Status = "ok"
	result.Summary = summary
	result.RawResult = raw
	result.DurationMS = time.Since(start).Milliseconds()
	return result
}

// fail creates an error result.
func (e *Executor) fail(start time.Time, result *RunResult, format string, args ...any) RunResult {
	result.Status = "error"
	result.Error = fmt.Sprintf(format, args...)
	result.DurationMS = time.Since(start).Milliseconds()
	return *result
}

// startTracker starts the tracker (locally or via protocol).
func (e *Executor) startTracker(ctx context.Context, udpRoutines, httpRoutines int, run BenchmarkRun) error {
	if e.tracker != nil {
		return e.startTrackerLocal(ctx, udpRoutines, httpRoutines, run)
	}
	return e.startTrackerRemote(udpRoutines, httpRoutines)
}

// startTrackerLocal handles local tracker startup.
func (e *Executor) startTrackerLocal(ctx context.Context, udpRoutines, httpRoutines int, run BenchmarkRun) error {
	if e.tracker.IsAlive() {
		slog.Info("  Stopping existing tracker...")
		if err := e.tracker.Stop(ctx); err != nil {
			slog.Warn("stop tracker", "error", err)
		}
		time.Sleep(time.Second)
	}

	slog.Info("  Clearing cache...")
	if err := e.tracker.ClearCache(); err != nil {
		slog.Warn("failed to clear cache", "error", err)
	}

	slog.Info(fmt.Sprintf("  Starting tracker (%s_routines=%d, other=%d)...",
		run.Mode, run.Routines, run.OtherRoutines))

	return e.tracker.Start(ctx, udpRoutines, httpRoutines)
}

// startTrackerRemote handles remote tracker startup via protocol.
func (e *Executor) startTrackerRemote(udpRoutines, httpRoutines int) error {
	slog.Info("  Requesting server START...")

	if err := e.conn.SendStart(udpRoutines, httpRoutines); err != nil {
		return fmt.Errorf("send START: %w", err)
	}

	e.conn.SetDeadline(time.Now().Add(readyTimeout))
	defer e.conn.SetDeadline(time.Time{})

	msgType, data, err := e.conn.RecvAny()
	if err != nil {
		return fmt.Errorf("recv response: %w", err)
	}

	switch msgType {
	case MsgTypeReady:
		slog.Info("  Server ready")
		return nil
	case MsgTypeError:
		var errMsg ErrorMessage
		json.Unmarshal(data, &errMsg)
		return fmt.Errorf("server error: %s", errMsg.ErrorMsg)
	default:
		return fmt.Errorf("unexpected response: %s", msgType)
	}
}

// stopTracker stops the tracker (locally or via protocol).
func (e *Executor) stopTracker(ctx context.Context) {
	if e.tracker != nil {
		slog.Info("  Stopping tracker...")
		if err := e.tracker.Stop(ctx); err != nil {
			slog.Warn("stop tracker", "error", err)
		}
	}
}

// notifyDone notifies the server that the benchmark is done (client mode only).
func (e *Executor) notifyDone(udpRoutines, httpRoutines int, resultFile string, success bool, errorMsg string) error {
	slog.Info("  Sending DONE to server...")

	if err := e.conn.SendDone(udpRoutines, httpRoutines, resultFile, success, errorMsg); err != nil {
		return fmt.Errorf("send DONE: %w", err)
	}

	e.conn.SetDeadline(time.Now().Add(stoppedTimeout))
	defer e.conn.SetDeadline(time.Time{})

	msgType, data, err := e.conn.RecvAny()
	if err != nil {
		return fmt.Errorf("recv STOPPED: %w", err)
	}

	switch msgType {
	case MsgTypeStopped:
		return nil
	case MsgTypeError:
		var errMsg ErrorMessage
		json.Unmarshal(data, &errMsg)
		return fmt.Errorf("server error on stop: %s", errMsg.ErrorMsg)
	default:
		return fmt.Errorf("unexpected response: %s", msgType)
	}
}
