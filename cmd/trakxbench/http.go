package main

import (
	"context"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"strings"
	"time"
)

type httpWorker struct {
	addr    string
	timeout time.Duration
	host    string
}

func newHTTPWorker(addr string, timeout time.Duration, hostHeader string) *httpWorker {
	return &httpWorker{
		addr:    addr,
		timeout: timeout,
		host:    hostHeader,
	}
}

func (w *httpWorker) doRequest(payload string) error {
	dialer := net.Dialer{Timeout: w.timeout}
	conn, err := dialer.Dial("tcp", w.addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(w.timeout))
	if _, err := io.WriteString(conn, payload); err != nil {
		return err
	}
	buf := make([]byte, 4096)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func buildHTTPAnnounce(hash, peer, host string, port int, left int64, numwant int, compact bool) string {
	var b strings.Builder
	b.Grow(256)
	b.WriteString("GET /announce?info_hash=")
	b.WriteString(hash)
	b.WriteString("&peer_id=")
	b.WriteString(peer)
	b.WriteString("&port=")
	b.WriteString(fmt.Sprintf("%d", port))
	b.WriteString("&uploaded=0&downloaded=0&left=")
	b.WriteString(fmt.Sprintf("%d", left))
	if compact {
		b.WriteString("&compact=1")
	}
	if numwant >= 0 {
		b.WriteString("&numwant=")
		b.WriteString(fmt.Sprintf("%d", numwant))
	}
	b.WriteString(" HTTP/1.1\r\nHost: ")
	b.WriteString(host)
	b.WriteString("\r\n\r\n")
	return b.String()
}

func buildHTTPScrape(hashes []string, host string) string {
	var b strings.Builder
	b.Grow(256)
	b.WriteString("GET /scrape?")
	for i, h := range hashes {
		if i > 0 {
			b.WriteString("&")
		}
		b.WriteString("info_hash=")
		b.WriteString(h)
	}
	b.WriteString(" HTTP/1.1\r\nHost: ")
	b.WriteString(host)
	b.WriteString("\r\n\r\n")
	return b.String()
}

type httpBenchWorker struct {
	id     int
	client *httpWorker
}

func (w *httpBenchWorker) run(ctx context.Context, cfg config, ds *dataset, limiter *rateLimiter) *workerMetrics {
	metrics := newWorkerMetrics()
	rng := mrand.New(mrand.NewSource(cfg.rngSeed + int64(w.id*3571)))
	peerIdx := w.id % len(ds.peers)
	peer := ds.encodedPeers[peerIdx]
	left := int64(1000)
	if rng.Float64() < cfg.seedFraction {
		left = 0
	}
	var hashIndexes []int
	for i := 0; i < cfg.torrentsPerPeer; i++ {
		hashIndexes = append(hashIndexes, rng.Intn(len(ds.torrents)))
	}
	for {
		select {
		case <-ctx.Done():
			return metrics
		default:
		}
		if limiter != nil {
			select {
			case <-ctx.Done():
				return metrics
			case <-limiter.tokens:
			}
		}
		reqType := requestAnnounce
		if rng.Float64() < cfg.scrapeRatio {
			reqType = requestScrape
		}
		start := time.Now()
		var err error
		if reqType == requestAnnounce {
			numwant := cfg.pickNumwant(rng)
			hash := ds.encodedHashes[hashIndexes[rng.Intn(len(hashIndexes))]]
			payload := buildHTTPAnnounce(hash, peer, cfg.httpHostHeader, defaultPort, left, numwant, cfg.compact)
			err = w.client.doRequest(payload)
		} else {
			hashes := make([]string, cfg.scrapeHashes)
			for i := 0; i < cfg.scrapeHashes; i++ {
				hashes[i] = ds.encodedHashes[rng.Intn(len(ds.encodedHashes))]
			}
			payload := buildHTTPScrape(hashes, cfg.httpHostHeader)
			err = w.client.doRequest(payload)
		}
		lat := time.Since(start)
		metrics.counts[reqType]++
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				metrics.timeouts++
			} else {
				metrics.errors++
			}
			continue
		}
		metrics.success++
		metrics.hist.observe(lat)
	}
}
