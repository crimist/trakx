package main

import (
	"flag"
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
	compact         bool
	torrents        int
	torrentsPerPeer int
	seedTorrents    int
	seedPeersPerTor int
	seedOnly        bool
	seedFraction    float64
	timeout         time.Duration
	rngSeed         int64
	outPath         string
	label           string
	udpConnRefresh  time.Duration
	httpHostHeader  string
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
	flag.BoolVar(&cfg.compact, "compact", true, "use compact responses (http)")
	flag.IntVar(&cfg.torrents, "torrents", 2048, "torrents in active dataset")
	flag.IntVar(&cfg.torrentsPerPeer, "torrents-per-peer", 4, "torrents per peer")
	flag.IntVar(&cfg.seedTorrents, "seed-torrents", 0, "seed torrents before run (0 to disable)")
	flag.IntVar(&cfg.seedPeersPerTor, "seed-peers-per-torrent", 0, "peers per torrent during seed phase")
	flag.BoolVar(&cfg.seedOnly, "seed-only", false, "run seed phase only")
	flag.Float64Var(&cfg.seedFraction, "seed-fraction", 0.2, "fraction of peers that are seeds (left=0)")
	flag.DurationVar(&cfg.timeout, "timeout", defaultTimeout, "per-request timeout")
	flag.DurationVar(&cfg.udpConnRefresh, "udp-conn-refresh", 10*time.Minute, "refresh UDP connection ID interval")
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
