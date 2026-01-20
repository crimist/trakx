package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "run":
		if err := runLocal(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "server":
		if err := runServer(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "client":
		if err := runClient(args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "version", "--version", "-v":
		fmt.Printf("orchestrate %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(2)
	}
}

func runLocal(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fs := flag.NewFlagSet("run", flag.ExitOnError)
	scenarioPath := fs.String("scenario", "", "scenario YAML file")
	routinesStr := fs.String("routines", "4,8,16,32", "comma-separated goroutine counts")
	mode := fs.String("mode", "udp", "protocol mode: udp, http")
	durationStr := fs.String("duration", "2m", "benchmark duration")
	seed := fs.Int("seed", 2048, "seed torrents")
	concurrency := fs.Int("concurrency", 64, "benchmark concurrency")
	rate := fs.Int("rate", 0, "target rate (0 = max)")
	outputDir := fs.String("output", "", "output directory")
	configPath := fs.String("config", "configs/bench/trakx.yaml", "trakx config file")
	httpAddr := fs.String("http-addr", "127.0.0.1:1337", "tracker HTTP address")
	udpAddr := fs.String("udp-addr", "127.0.0.1:1337", "tracker UDP address")
	target := fs.String("target", "127.0.0.1:1337", "benchmark target address")
	quiet := fs.Bool("quiet", false, "suppress progress output")
	pause := fs.Bool("pause", false, "pause between runs")
	readyTimeout := fs.Duration("ready-timeout", 30*time.Second, "tracker readiness timeout")

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

	scenario, err := LoadOrBuildScenario(BenchConfig{
		RepoRoot:     repoRoot,
		ScenarioPath: *scenarioPath,
		RoutinesStr:  *routinesStr,
		Mode:         *mode,
		DurationStr:  *durationStr,
		Seed:         *seed,
		Concurrency:  *concurrency,
		Rate:         *rate,
	})
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	runs, err := scenario.Expand()
	if err != nil {
		return fmt.Errorf("expand scenario: %w", err)
	}

	if len(runs) == 0 {
		return fmt.Errorf("no benchmark runs defined")
	}

	*outputDir, err = SetupOutputDir(repoRoot, *outputDir, *mode)
	if err != nil {
		return err
	}

	slog.Info("Building trakx...")
	trakxBin := filepath.Join(repoRoot, "bench", "bin", "trakx")
	if err := EnsureTrakxBinary(repoRoot, trakxBin); err != nil {
		return fmt.Errorf("build trakx: %w", err)
	}

	slog.Info("Building bench...")
	benchBin := filepath.Join(repoRoot, "bench", "bin", "bench")
	if err := EnsureBenchBinary(repoRoot, benchBin); err != nil {
		return fmt.Errorf("build bench: %w", err)
	}

	tracker := NewTracker(trakxBin, *configPath, *httpAddr, *udpAddr, *readyTimeout)
	executor := NewLocalExecutor(tracker, benchBin, *target, *outputDir)
	collector := NewResultCollector(*outputDir, scenario.Name, *mode)

	slog.Info(fmt.Sprintf("\nRunning %d benchmark(s)...\n", len(runs)))

	for i, run := range runs {
		slog.Info(fmt.Sprintf("[%d/%d] %s (routines=%d)", i+1, len(runs), run.Name, run.Routines))

		if *pause && i > 0 {
			fmt.Fprint(os.Stderr, "Press Enter to continue...")
			if err := WaitForInput(); err != nil {
				return fmt.Errorf("wait for input: %w", err)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result := executor.Run(ctx, run)
		collector.AddResult(result)

		if result.Summary != nil {
			slog.Info(fmt.Sprintf("  -> %.0f req/s | p50: %s | p90: %s | p95: %s | p99: %s",
				result.Summary.Throughput, result.Summary.P50, result.Summary.P90, result.Summary.P95, result.Summary.P99))
		} else if result.Error != "" {
			slog.Error(fmt.Sprintf("  -> %s", result.Error))
		}
	}

	return FinalizeResults(collector)
}

func runClient(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fs := flag.NewFlagSet("client", flag.ExitOnError)
	serverAddr := fs.String("server", "", "server address (required)")
	scenarioPath := fs.String("scenario", "", "scenario YAML file")
	routinesStr := fs.String("routines", "4,8,16,32", "comma-separated goroutine counts")
	mode := fs.String("mode", "udp", "protocol mode: udp, http")
	durationStr := fs.String("duration", "2m", "benchmark duration")
	seed := fs.Int("seed", 2048, "seed torrents")
	concurrency := fs.Int("concurrency", 64, "benchmark concurrency")
	rate := fs.Int("rate", 0, "target rate (0 = max)")
	target := fs.String("target", "", "tracker target address (default: server:1337)")
	outputDir := fs.String("output", "", "output directory")
	quiet := fs.Bool("quiet", false, "suppress progress output")
	pause := fs.Bool("pause", false, "pause between runs")
	statsURL := fs.String("stats", "", "stats URL (optional)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *serverAddr == "" {
		return fmt.Errorf("--server is required")
	}

	SetupLogging(*quiet)

	if *target == "" {
		host, _, err := net.SplitHostPort(*serverAddr)
		if err != nil {
			host = *serverAddr
		}
		*target = fmt.Sprintf("%s:1337", host)
	}

	if *statsURL == "" {
		host, _, err := net.SplitHostPort(*serverAddr)
		if err != nil {
			host = *serverAddr
		}
		*statsURL = fmt.Sprintf("http://%s:1337/stats", host)
	}

	repoRoot, err := FindRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	scenario, err := LoadOrBuildScenario(BenchConfig{
		RepoRoot:     repoRoot,
		ScenarioPath: *scenarioPath,
		RoutinesStr:  *routinesStr,
		Mode:         *mode,
		DurationStr:  *durationStr,
		Seed:         *seed,
		Concurrency:  *concurrency,
		Rate:         *rate,
	})
	if err != nil {
		return fmt.Errorf("load scenario: %w", err)
	}

	runs, err := scenario.Expand()
	if err != nil {
		return fmt.Errorf("expand scenario: %w", err)
	}

	if len(runs) == 0 {
		return fmt.Errorf("no benchmark runs defined")
	}

	*outputDir, err = SetupOutputDir(repoRoot, *outputDir, *mode)
	if err != nil {
		return err
	}

	slog.Info("Building bench...")
	benchBin := filepath.Join(repoRoot, "bench", "bin", "bench")
	if err := EnsureBenchBinary(repoRoot, benchBin); err != nil {
		return fmt.Errorf("build bench: %w", err)
	}

	slog.Info(fmt.Sprintf("Connecting to server at %s...", *serverAddr))
	netConn, err := net.DialTimeout("tcp", *serverAddr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("connect to server: %w", err)
	}
	defer netConn.Close()

	conn := NewConn(netConn)
	executor := NewClientExecutor(conn, benchBin, *target, *statsURL, *outputDir)
	collector := NewResultCollector(*outputDir, scenario.Name, *mode)

	slog.Info(fmt.Sprintf("Connected. Running %d benchmark(s)...\n", len(runs)))

	for i, run := range runs {
		slog.Info(fmt.Sprintf("[%d/%d] %s (routines=%d)", i+1, len(runs), run.Name, run.Routines))

		if *pause && i > 0 {
			fmt.Fprint(os.Stderr, "Press Enter to continue...")
			if err := WaitForInput(); err != nil {
				return fmt.Errorf("wait for input: %w", err)
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result := executor.Run(ctx, run)
		collector.AddResult(result)

		if result.Summary != nil {
			slog.Info(fmt.Sprintf("  -> %.0f req/s | p50: %s | p90: %s | p95: %s | p99: %s",
				result.Summary.Throughput, result.Summary.P50, result.Summary.P90, result.Summary.P95, result.Summary.P99))
		} else if result.Error != "" {
			slog.Error(fmt.Sprintf("  -> %s", result.Error))
		}
	}

	return FinalizeResults(collector)
}

func printUsage() {
	fmt.Print(`orchestrate - Benchmark orchestration for trakx

Usage:
  orchestrate <command> [options]

Commands:
  run       Run benchmarks locally (single machine)
  server    Start distributed mode server (runs on tracker machine)
  client    Start distributed mode client (runs on benchmark machine)
  version   Print version information
  help      Show this help message

Local Mode:
  orchestrate run [options]
    --scenario <file>     Scenario YAML file (default: inline sweep)
    --routines <list>     Comma-separated goroutine counts (default: 4,8,16,32)
    --mode <mode>         Protocol mode: udp, http (default: udp)
    --duration <dur>      Benchmark duration (default: 2m)
    --seed <n>            Seed torrents (default: 2048)
    --output <dir>        Output directory (default: ./results/<timestamp>)
    --config <file>       Trakx config file (default: configs/bench/trakx.yaml)
    --quiet               Suppress progress output
    --pause               Pause between runs for manual inspection

Distributed Mode (Server):
  orchestrate server [options]
    --listen <addr>       Listen address (default: 0.0.0.0:9077)
    --config <file>       Trakx config file (default: configs/bench/trakx.yaml)
    --http-addr <addr>    Tracker HTTP address for readiness (default: 127.0.0.1:1337)
    --udp-addr <addr>     Tracker UDP address for readiness (default: 127.0.0.1:1337)
    --ready-timeout <dur> Readiness timeout (default: 30s)

Distributed Mode (Client):
  orchestrate client [options]
    --server <addr>       Server address (required)
    --scenario <file>     Scenario YAML file (default: inline sweep)
    --routines <list>     Comma-separated goroutine counts (default: 4,8,16,32)
    --mode <mode>         Protocol mode: udp, http (default: udp)
    --duration <dur>      Benchmark duration (default: 2m)
    --seed <n>            Seed torrents (default: 2048)
    --target <addr>       Tracker address for benchmarking (default: <server>:1337)
    --output <dir>        Output directory (default: ./results/<timestamp>)
    --quiet               Suppress progress output
    --pause               Pause between runs

Examples:
  # Local sweep through goroutine counts
  orchestrate run --routines 4,8,16,32 --mode udp --duration 2m

  # Local with custom scenario file
  orchestrate run --scenario configs/bench/scenarios/comprehensive.yaml

  # Distributed: start server on tracker machine
  orchestrate server --listen 0.0.0.0:9077 --config configs/bench/trakx.yaml

  # Distributed: run client from benchmark machine
  orchestrate client --server 192.168.1.100:9077 --routines 4,8,16,32

`)
}
