package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/klauspost/cpuid/v2"
)

// Hardcoded realistic defaults for BitTorrent workloads.
const (
	defaultWarmup      = 10 * time.Second
	defaultTimeout     = 3 * time.Second
	defaultConnRefresh = 10 * time.Minute

	// Request distribution (realistic BitTorrent)
	defaultScrapeRatio     = 0.03 // 3% scrapes
	defaultScrapeHashes    = 5
	defaultTorrentsPerPeer = 4
	defaultCompact         = true

	// Seed distribution (lognormal - realistic)
	defaultSeedMedian   = 30.0
	defaultSeedSigma    = 1.2
	defaultSeedMin      = 10
	defaultSeedMax      = 1000
	defaultSeedFraction = 0.2 // 20% seeders

	defaultPort = 6881

	// RNG seed offsets for reproducible but distinct sequences per worker type.
	rngOffsetUDP  = 7919
	rngOffsetHTTP = 3571
	rngOffsetSeed = 11
)

// Hardcoded numwant distribution: 20 (90%), 30 (8%), 50 (2%)
var defaultNumwantWeights = []weightedChoice{
	{value: 20, cumulative: 0.90},
	{value: 30, cumulative: 0.98},
	{value: 50, cumulative: 1.00},
}

type config struct {
	mode        string
	target      string
	statsURL    string
	duration    time.Duration
	rate        int
	concurrency int
	seed        int
	rngSeed     int64
	outPath     string
	quiet       bool
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.mode, "mode", "udp", "protocol: udp, http, or both")
	flag.StringVar(&cfg.target, "target", "127.0.0.1:1337", "tracker address host:port")
	flag.DurationVar(&cfg.duration, "duration", 2*time.Minute, "measurement duration")
	flag.IntVar(&cfg.rate, "rate", 0, "requests/sec (0 = max throughput)")
	flag.IntVar(&cfg.concurrency, "concurrency", 64, "worker count")
	flag.IntVar(&cfg.seed, "seed", 0, "pre-seed with N torrents (0 = disabled)")
	flag.Int64Var(&cfg.rngSeed, "rng-seed", 0, "RNG seed (0 = random)")
	flag.StringVar(&cfg.outPath, "out", "", "output file (default stdout)")
	flag.StringVar(&cfg.statsURL, "stats", "", "expvar stats URL (optional)")
	flag.BoolVar(&cfg.quiet, "quiet", false, "suppress progress output")
	flag.Parse()
	return cfg
}

type result struct {
	GeneratedAt string                 `json:"generated_at"`
	Mode        string                 `json:"mode"`
	Target      string                 `json:"target"`
	Config      map[string]interface{} `json:"config"`
	Metrics     map[string]metricView  `json:"metrics"`
	Stats       map[string]interface{} `json:"stats,omitempty"`
	Runtime     map[string]interface{} `json:"runtime"`
	Warnings    []string               `json:"warnings,omitempty"`
}

func main() {
	cfg := parseFlags()
	if cfg.mode != "udp" && cfg.mode != "http" && cfg.mode != "both" {
		fmt.Fprintln(os.Stderr, "mode must be udp, http, or both")
		os.Exit(2)
	}
	if cfg.concurrency <= 0 {
		fmt.Fprintln(os.Stderr, "concurrency must be > 0")
		os.Exit(2)
	}

	seed := cfg.rngSeed
	if seed == 0 {
		seed = cryptoSeed()
	}
	cfg.rngSeed = seed
	rng := newRand(seed)

	torrents := cfg.seed
	if torrents <= 0 {
		torrents = 2048
	}
	ds := newDataset(rng, torrents, cfg.concurrency)

	if cfg.seed > 0 {
		seedCtx, cancel := context.WithTimeout(context.Background(), defaultWarmup+cfg.duration)
		if !cfg.quiet {
			fmt.Fprintf(os.Stderr, "seeding %d torrents...\n", cfg.seed)
		}
		if err := seedPhase(seedCtx, cfg, ds); err != nil {
			cancel()
			fmt.Fprintf(os.Stderr, "seed phase failed: %v\n", err)
			os.Exit(1)
		}
		cancel()
		if !cfg.quiet {
			fmt.Fprintln(os.Stderr, "seeding complete")
		}
	}

	results := make(map[string]metricView)
	var warnings []string

	if cfg.mode == "both" {
		cfgUDP := cfg
		cfgUDP.mode = "udp"
		if !cfg.quiet {
			fmt.Fprintln(os.Stderr, "benchmarking UDP...")
		}
		udpResult, warn, err := runBenchmark(cfgUDP, ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "udp benchmark failed: %v\n", err)
			os.Exit(1)
		}
		results["udp"] = udpResult
		warnings = append(warnings, warn...)

		cfgHTTP := cfg
		cfgHTTP.mode = "http"
		if !cfg.quiet {
			fmt.Fprintln(os.Stderr, "benchmarking HTTP...")
		}
		httpResult, warn, err := runBenchmark(cfgHTTP, ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "http benchmark failed: %v\n", err)
			os.Exit(1)
		}
		results["http"] = httpResult
		warnings = append(warnings, warn...)
	} else {
		if !cfg.quiet {
			fmt.Fprintf(os.Stderr, "benchmarking %s...\n", cfg.mode)
		}
		res, warn, err := runBenchmark(cfg, ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "benchmark failed: %v\n", err)
			os.Exit(1)
		}
		results[cfg.mode] = res
		warnings = append(warnings, warn...)
	}

	stats := map[string]interface{}{}
	if cfg.statsURL != "" {
		if data, err := fetchStats(cfg.statsURL, defaultTimeout); err == nil {
			stats = data
		} else {
			warnings = append(warnings, fmt.Sprintf("stats fetch failed: %v", err))
		}
	}

	out := result{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Mode:        cfg.mode,
		Target:      cfg.target,
		Config: map[string]interface{}{
			"duration":    cfg.duration.String(),
			"rate":        cfg.rate,
			"concurrency": cfg.concurrency,
			"seed":        cfg.seed,
			"rng_seed":    cfg.rngSeed,
		},
		Metrics:  results,
		Stats:    stats,
		Runtime: map[string]interface{}{
			"go":         runtime.Version(),
			"gomaxprocs": runtime.GOMAXPROCS(0),
			"cpu_model":  cpuid.CPU.BrandName,
		},
		Warnings: warnings,
	}

	var w io.Writer = os.Stdout
	if cfg.outPath != "" {
		f, err := os.Create(cfg.outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		w = f
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode results: %v\n", err)
		os.Exit(1)
	}
	if !cfg.quiet {
		fmt.Fprintln(os.Stderr, "done")
	}
}
