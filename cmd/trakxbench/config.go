package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

const (
	defaultScrapeRatio  = 0.03
	defaultTimeout      = 3 * time.Second
	defaultWarmup       = 20 * time.Second
	defaultDuration     = 2 * time.Minute
	defaultScrapeHashes = 5
	defaultPort         = 6881
)

type config struct {
	mode            string
	udpAddr         string
	httpAddr        string
	statsURL        string
	duration        time.Duration
	warmup          time.Duration
	rate            int
	concurrency     int
	scrapeRatio     float64
	scrapeHashes    int
	numwant         int
	numwantDist     string
	numwantWeights  string
	compact         bool
	torrents        int
	torrentsPerPeer int
	seedTorrents    int
	seedPeersPerTor int
	seedPeersDist   string
	seedPeersMedian float64
	seedPeersMean   float64
	seedPeersSigma  float64
	seedPeersMin    int
	seedPeersMax    int
	seedOnly        bool
	seedFraction    float64
	timeout         time.Duration
	rngSeed         int64
	outPath         string
	label           string
	udpConnRefresh  time.Duration
	httpHostHeader  string
	numwantChoices  []weightedChoice
}

func parseFlags() config {
	var cfg config
	flag.StringVar(&cfg.mode, "mode", "udp", "benchmark mode: udp, http, or both")
	flag.StringVar(&cfg.udpAddr, "udp", "127.0.0.1:1337", "UDP tracker address host:port")
	flag.StringVar(&cfg.httpAddr, "http", "127.0.0.1:1337", "HTTP tracker address host:port")
	flag.StringVar(&cfg.statsURL, "stats", "", "expvar stats URL (optional)")
	flag.DurationVar(&cfg.duration, "duration", defaultDuration, "measurement duration")
	flag.DurationVar(&cfg.warmup, "warmup", defaultWarmup, "warmup duration before measurement")
	flag.IntVar(&cfg.rate, "rate", 0, "target request rate per second (0 for max)")
	flag.IntVar(&cfg.concurrency, "concurrency", 64, "number of client workers")
	flag.Float64Var(&cfg.scrapeRatio, "scrape-ratio", defaultScrapeRatio, "scrape ratio (0-1)")
	flag.IntVar(&cfg.scrapeHashes, "scrape-hashes", defaultScrapeHashes, "info_hashes per scrape request")
	flag.IntVar(&cfg.numwant, "numwant", -1, "numwant parameter (-1 to omit)")
	flag.StringVar(&cfg.numwantDist, "numwant-dist", "fixed", "numwant distribution: fixed or weighted")
	flag.StringVar(&cfg.numwantWeights, "numwant-weights", "20:0.90,30:0.08,50:0.02", "weighted numwant distribution")
	flag.BoolVar(&cfg.compact, "compact", true, "use compact responses (http)")
	flag.IntVar(&cfg.torrents, "torrents", 2048, "torrents in active dataset")
	flag.IntVar(&cfg.torrentsPerPeer, "torrents-per-peer", 4, "torrents per peer")
	flag.IntVar(&cfg.seedTorrents, "seed-torrents", 0, "seed torrents before run (0 to disable)")
	flag.IntVar(&cfg.seedPeersPerTor, "seed-peers-per-torrent", 0, "peers per torrent during seed phase")
	flag.StringVar(&cfg.seedPeersDist, "seed-peers-dist", "fixed", "seed peers per torrent distribution: fixed, normal, lognormal")
	flag.Float64Var(&cfg.seedPeersMedian, "seed-peers-median", 30, "lognormal median peers per torrent")
	flag.Float64Var(&cfg.seedPeersMean, "seed-peers-mean", 30, "normal mean peers per torrent")
	flag.Float64Var(&cfg.seedPeersSigma, "seed-peers-sigma", 1.2, "distribution sigma for normal/lognormal")
	flag.IntVar(&cfg.seedPeersMin, "seed-peers-min", 10, "minimum peers per torrent when seeding")
	flag.IntVar(&cfg.seedPeersMax, "seed-peers-max", 1000, "maximum peers per torrent when seeding")
	flag.BoolVar(&cfg.seedOnly, "seed-only", false, "run seed phase only")
	flag.Float64Var(&cfg.seedFraction, "seed-fraction", 0.2, "fraction of peers that are seeds (left=0)")
	flag.DurationVar(&cfg.timeout, "timeout", defaultTimeout, "per-request timeout")
	flag.DurationVar(&cfg.udpConnRefresh, "udp-conn-refresh", 10*time.Minute, "refresh UDP connection ID interval")
	flag.Int64Var(&cfg.rngSeed, "rng-seed", 0, "rng seed (0 for random)")
	flag.StringVar(&cfg.outPath, "out", "", "write results JSON to file (default stdout)")
	flag.StringVar(&cfg.label, "label", "", "label for this run (optional)")
	flag.Parse()

	if cfg.scrapeRatio < 0 {
		cfg.scrapeRatio = 0
	}
	if cfg.scrapeRatio > 1 {
		cfg.scrapeRatio = 1
	}
	if cfg.torrentsPerPeer <= 0 {
		cfg.torrentsPerPeer = 1
	}
	cfg.httpHostHeader = cfg.httpAddr
	if strings.Contains(cfg.httpAddr, ":") {
		cfg.httpHostHeader = cfg.httpAddr
	}
	return cfg
}

func (cfg *config) finalize() error {
	cfg.numwantDist = strings.ToLower(cfg.numwantDist)
	cfg.seedPeersDist = strings.ToLower(cfg.seedPeersDist)

	switch cfg.numwantDist {
	case "fixed":
	case "weighted":
		choices, err := parseWeightedChoices(cfg.numwantWeights)
		if err != nil {
			return err
		}
		cfg.numwantChoices = choices
	default:
		return fmt.Errorf("invalid numwant-dist: %s", cfg.numwantDist)
	}

	switch cfg.seedPeersDist {
	case "fixed", "normal", "lognormal":
	default:
		return fmt.Errorf("invalid seed-peers-dist: %s", cfg.seedPeersDist)
	}

	if cfg.seedPeersMin < 0 {
		cfg.seedPeersMin = 0
	}
	if cfg.seedPeersMax < 0 {
		cfg.seedPeersMax = 0
	}
	if cfg.seedPeersMax > 0 && cfg.seedPeersMin > cfg.seedPeersMax {
		return fmt.Errorf("seed-peers-min > seed-peers-max")
	}
	return nil
}
