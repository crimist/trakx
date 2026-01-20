package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type worker interface {
	run(ctx context.Context, cfg config, ds *dataset, limiter *rateLimiter) *workerMetrics
}

func runPhase(ctx context.Context, cfg config, ds *dataset, workers []worker, limiter *rateLimiter) *benchMetrics {
	var wg sync.WaitGroup
	metrics := newBenchMetrics()
	results := make(chan *workerMetrics, len(workers))
	for _, w := range workers {
		wg.Add(1)
		go func(w worker) {
			defer wg.Done()
			results <- w.run(ctx, cfg, ds, limiter)
		}(w)
	}
	wg.Wait()
	close(results)
	for res := range results {
		metrics.merge(res)
	}
	return metrics
}

func runBenchmark(cfg config, ds *dataset) (metricView, []string, error) {
	var warnings []string
	workers, err := buildWorkers(cfg)
	if err != nil {
		return metricView{}, warnings, err
	}
	defer closeWorkers(workers)

	if defaultWarmup > 0 {
		warmCtx, cancel := context.WithTimeout(context.Background(), defaultWarmup)
		limiter := newRateLimiter(warmCtx, cfg.rate)
		_ = runPhase(warmCtx, cfg, ds, workers, limiter)
		cancel()
	}

	runCtx, cancel := context.WithTimeout(context.Background(), cfg.duration)
	defer cancel()
	limiter := newRateLimiter(runCtx, cfg.rate)
	start := time.Now()
	metrics := runPhase(runCtx, cfg, ds, workers, limiter)
	elapsed := time.Since(start)

	if elapsed < cfg.duration-(250*time.Millisecond) {
		warnings = append(warnings, "benchmark completed earlier than expected")
	}
	view := summarize(metrics, elapsed)
	return view, warnings, nil
}

func buildWorkers(cfg config) ([]worker, error) {
	switch cfg.mode {
	case "udp":
		addr, err := net.ResolveUDPAddr("udp", cfg.target)
		if err != nil {
			return nil, err
		}
		workers := make([]worker, 0, cfg.concurrency)
		for i := 0; i < cfg.concurrency; i++ {
			client, err := newUDPWorker(addr, defaultTimeout, defaultConnRefresh)
			if err != nil {
				return nil, err
			}
			workers = append(workers, &udpBenchWorker{id: i, client: client})
		}
		return workers, nil
	case "http":
		workers := make([]worker, 0, cfg.concurrency)
		for i := 0; i < cfg.concurrency; i++ {
			workers = append(workers, &httpBenchWorker{id: i, client: newHTTPWorker(cfg.target, defaultTimeout, cfg.target)})
		}
		return workers, nil
	default:
		return nil, fmt.Errorf("unknown mode: %s", cfg.mode)
	}
}

func closeWorkers(workers []worker) {
	for _, w := range workers {
		if uw, ok := w.(*udpBenchWorker); ok {
			uw.client.close()
		}
	}
}
