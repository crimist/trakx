package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// Error codes for protocol error messages.
const (
	ErrCodeInternal     = 1 // Internal server error
	ErrCodeTrackerStart = 2 // Tracker failed to start
	ErrCodeUnknownMsg   = 3 // Unknown message type
	ErrCodeInvalidField = 4 // Missing or invalid message field
)

// Timeout constants for server operations.
const (
	serverContextCheckInterval = 5 * time.Second
	serverTrackerStopWait      = 10 * time.Second
)

func runServer(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fs := flag.NewFlagSet("server", flag.ExitOnError)
	listenAddr := fs.String("listen", "0.0.0.0:9077", "listen address")
	configPath := fs.String("config", "configs/bench/trakx.yaml", "trakx config file")
	httpAddr := fs.String("http-addr", "127.0.0.1:1337", "tracker HTTP address for readiness")
	udpAddr := fs.String("udp-addr", "127.0.0.1:1337", "tracker UDP address for readiness")
	readyTimeout := fs.Duration("ready-timeout", 30*time.Second, "tracker readiness timeout")
	quiet := fs.Bool("quiet", false, "suppress progress output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	SetupLogging(*quiet)

	repoRoot, err := FindRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	if !filepath.IsAbs(*configPath) {
		*configPath = filepath.Join(repoRoot, *configPath)
	}

	slog.Info("Building trakx...")
	trakxBin := filepath.Join(repoRoot, "bench", "bin", "trakx")
	if err := EnsureTrakxBinary(repoRoot, trakxBin); err != nil {
		return fmt.Errorf("build trakx: %w", err)
	}

	tracker := NewTracker(trakxBin, *configPath, *httpAddr, *udpAddr, *readyTimeout)

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("\nReceived shutdown signal...")
		listener.Close()
		cancel()
	}()

	slog.Info(fmt.Sprintf("Listening on %s", *listenAddr))
	slog.Info("Waiting for client connection...")

	for {
		netConn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				if tracker.IsAlive() {
					slog.Info("Stopping tracker before shutdown...")
					if err := tracker.Stop(context.Background()); err != nil {
						slog.Warn("stop tracker", "error", err)
					}
				}
				slog.Info("Server shutdown complete")
				return nil
			default:
				slog.Warn("Accept error", "error", err)
				continue
			}
		}

		slog.Info(fmt.Sprintf("Client connected from %s", netConn.RemoteAddr()))

		if err := handleClient(ctx, netConn, tracker, *httpAddr, *udpAddr); err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, context.Canceled) {
				slog.Error("Client error", "error", err)
			}
		}

		slog.Info("Client disconnected")
		slog.Info("Waiting for client connection...")
	}
}

func handleClient(ctx context.Context, netConn net.Conn, tracker *Tracker, httpAddr, udpAddr string) error {
	conn := NewConn(netConn)
	defer conn.Close()

	var currentUDP, currentHTTP int

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		conn.SetDeadline(time.Now().Add(serverContextCheckInterval))
		msgType, data, err := conn.RecvAny()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return err
		}
		conn.SetDeadline(time.Time{})

		switch msgType {
		case MsgTypeStart:
			var msg StartMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				conn.SendError(ErrCodeInvalidField, fmt.Sprintf("invalid START message: %v", err), true)
				continue
			}

			slog.Info(fmt.Sprintf("[%s] START: udp_routines=%d, http_routines=%d",
				time.Now().Format("15:04:05"), msg.UDPRoutines, msg.HTTPRoutines))

			if tracker.IsAlive() {
				slog.Info("  Stopping existing tracker...")
				deadline := time.Now().Add(serverTrackerStopWait)
				for time.Now().Before(deadline) && tracker.IsAlive() {
					tracker.Stop(ctx)
					time.Sleep(200 * time.Millisecond)
				}
				if tracker.IsAlive() {
					conn.SendError(ErrCodeInternal, "tracker still running after stop", false)
					continue
				}
			}

			slog.Info("  Starting tracker...")
			if err := tracker.Restart(ctx, msg.UDPRoutines, msg.HTTPRoutines); err != nil {
				slog.Error("start tracker failed", "error", err)
				conn.SendError(ErrCodeTrackerStart, fmt.Sprintf("start tracker: %v", err), true)
				continue
			}

			currentUDP = msg.UDPRoutines
			currentHTTP = msg.HTTPRoutines

			slog.Info("  Tracker ready")
			conn.SendReady(msg.UDPRoutines, msg.HTTPRoutines, httpAddr, udpAddr)

		case MsgTypeDone:
			var msg DoneMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				slog.Warn("invalid DONE message", "error", err)
				conn.SendError(ErrCodeInvalidField, fmt.Sprintf("invalid DONE message: %v", err), true)
				continue
			}

			status := "OK"
			if !msg.Success {
				status = "FAILED"
			}
			slog.Info(fmt.Sprintf("[%s] DONE: udp_routines=%d, http_routines=%d, status=%s",
				time.Now().Format("15:04:05"), msg.UDPRoutines, msg.HTTPRoutines, status))

			slog.Info("  Stopping tracker...")
			if err := tracker.Stop(ctx); err != nil {
				slog.Warn("stop tracker", "error", err)
			}
			time.Sleep(time.Second)

			conn.SendStopped(currentUDP, currentHTTP)
			currentUDP = 0
			currentHTTP = 0

		case MsgTypePing:
			conn.SendPong()

		case MsgTypeGetState:
			status := "idle"
			if tracker.IsAlive() {
				status = "running"
			}
			conn.SendState(status, 0)

		default:
			slog.Warn(fmt.Sprintf("Unknown message type: %s", msgType))
			conn.SendError(ErrCodeUnknownMsg, fmt.Sprintf("unknown message type: %s", msgType), true)
		}
	}
}
