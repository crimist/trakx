package main

import "time"

type metricView struct {
	Count    int64            `json:"count"`
	Success  int64            `json:"success"`
	Errors   int64            `json:"errors"`
	Timeouts int64            `json:"timeouts"`
	Rate     float64          `json:"rate_per_sec"`
	Latency  latencyView      `json:"latency"`
	ByType   map[string]int64 `json:"by_type,omitempty"`
}

type latencyView struct {
	Min  string `json:"min"`
	Max  string `json:"max"`
	Mean string `json:"mean"`
	P50  string `json:"p50"`
	P90  string `json:"p90"`
	P95  string `json:"p95"`
	P99  string `json:"p99"`
}

type requestType string

const (
	requestAnnounce requestType = "announce"
	requestScrape   requestType = "scrape"
)

type workerMetrics struct {
	counts   map[requestType]int64
	success  int64
	errors   int64
	timeouts int64
	hist     *latencyHistogram
}

func newWorkerMetrics() *workerMetrics {
	return &workerMetrics{
		counts: make(map[requestType]int64),
		hist:   newLatencyHistogram(100*time.Microsecond, 18),
	}
}

type benchMetrics struct {
	counts   map[requestType]int64
	success  int64
	errors   int64
	timeouts int64
	hist     *latencyHistogram
}

func newBenchMetrics() *benchMetrics {
	return &benchMetrics{
		counts: make(map[requestType]int64),
		hist:   newLatencyHistogram(100*time.Microsecond, 18),
	}
}

func (bm *benchMetrics) merge(worker *workerMetrics) {
	if worker == nil {
		return
	}
	for k, v := range worker.counts {
		bm.counts[k] += v
	}
	bm.success += worker.success
	bm.errors += worker.errors
	bm.timeouts += worker.timeouts
	bm.hist.merge(worker.hist)
}

func summarize(metrics *benchMetrics, duration time.Duration) metricView {
	count := int64(0)
	for _, v := range metrics.counts {
		count += v
	}
	rate := float64(metrics.success) / duration.Seconds()
	lat := latencyView{
		Min:  formatDuration(time.Duration(metrics.hist.min)),
		Max:  formatDuration(time.Duration(metrics.hist.max)),
		Mean: formatDuration(metrics.hist.mean()),
		P50:  metrics.hist.quantile(0.50).String(),
		P90:  metrics.hist.quantile(0.90).String(),
		P95:  metrics.hist.quantile(0.95).String(),
		P99:  metrics.hist.quantile(0.99).String(),
	}
	byType := make(map[string]int64)
	for k, v := range metrics.counts {
		byType[string(k)] = v
	}
	return metricView{
		Count:    count,
		Success:  metrics.success,
		Errors:   metrics.errors,
		Timeouts: metrics.timeouts,
		Rate:     rate,
		Latency:  lat,
		ByType:   byType,
	}
}
