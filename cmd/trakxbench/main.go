package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"
)

type result struct {
	GeneratedAt string                 `json:"generated_at"`
	Label       string                 `json:"label,omitempty"`
	Mode        string                 `json:"mode"`
	Target      map[string]string      `json:"target"`
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
	if cfg.torrents <= 0 {
		cfg.torrents = 1
	}
	ds := newDataset(rng, cfg.torrents, cfg.concurrency)

	if cfg.seedTorrents > 0 && cfg.seedPeersPerTor > 0 {
		seedCtx, cancel := context.WithTimeout(context.Background(), cfg.warmup+cfg.duration)
		if err := seedPhase(seedCtx, cfg, ds); err != nil {
			cancel()
			fmt.Fprintf(os.Stderr, "seed phase failed: %v\n", err)
			os.Exit(1)
		}
		cancel()
	}
	if cfg.seedOnly {
		return
	}

	results := make(map[string]metricView)
	var warnings []string

	if cfg.mode == "both" {
		cfgUDP := cfg
		cfgUDP.mode = "udp"
		udpResult, warn, err := runBenchmark(cfgUDP, ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "udp benchmark failed: %v\n", err)
			os.Exit(1)
		}
		results["udp"] = udpResult
		warnings = append(warnings, warn...)

		cfgHTTP := cfg
		cfgHTTP.mode = "http"
		httpResult, warn, err := runBenchmark(cfgHTTP, ds)
		if err != nil {
			fmt.Fprintf(os.Stderr, "http benchmark failed: %v\n", err)
			os.Exit(1)
		}
		results["http"] = httpResult
		warnings = append(warnings, warn...)
	} else {
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
		if data, err := fetchStats(cfg.statsURL, cfg.timeout); err == nil {
			stats = data
		} else {
			warnings = append(warnings, fmt.Sprintf("stats fetch failed: %v", err))
		}
	}

	out := result{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Label:       cfg.label,
		Mode:        cfg.mode,
		Target: map[string]string{
			"udp":  cfg.udpAddr,
			"http": cfg.httpAddr,
		},
		Config: map[string]interface{}{
			"duration":               cfg.duration.String(),
			"warmup":                 cfg.warmup.String(),
			"rate":                   cfg.rate,
			"concurrency":            cfg.concurrency,
			"scrape_ratio":           cfg.scrapeRatio,
			"scrape_hashes":          cfg.scrapeHashes,
			"numwant":                cfg.numwant,
			"compact":                cfg.compact,
			"torrents":               cfg.torrents,
			"torrents_per_peer":      cfg.torrentsPerPeer,
			"seed_torrents":          cfg.seedTorrents,
			"seed_peers_per_torrent": cfg.seedPeersPerTor,
			"seed_fraction":          cfg.seedFraction,
			"timeout":                cfg.timeout.String(),
			"udp_conn_refresh":       cfg.udpConnRefresh.String(),
			"rng_seed":               cfg.rngSeed,
		},
		Metrics:  results,
		Stats:    stats,
		Runtime:  map[string]interface{}{"go": runtime.Version(), "gomaxprocs": runtime.GOMAXPROCS(0)},
		Warnings: warnings,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if cfg.outPath != "" {
		f, err := os.Create(cfg.outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open output file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		enc = json.NewEncoder(f)
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode results: %v\n", err)
		os.Exit(1)
	}
}
