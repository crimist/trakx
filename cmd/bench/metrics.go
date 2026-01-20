package main

import (
	"math"
	"sort"
	"time"
)

// Latency histogram.

type latencyHistogram struct {
	bounds []time.Duration
	counts []int64
	sum    int64
	min    int64
	max    int64
}

func newLatencyHistogram(start time.Duration, buckets int) *latencyHistogram {
	if buckets < 2 {
		buckets = 2
	}
	bounds := make([]time.Duration, buckets)
	val := start
	for i := 0; i < buckets; i++ {
		bounds[i] = val
		val *= 2
	}
	return &latencyHistogram{
		bounds: bounds,
		counts: make([]int64, buckets+1),
		min:    math.MaxInt64,
	}
}

func (h *latencyHistogram) observe(d time.Duration) {
	if d < 0 {
		return
	}
	ns := int64(d)
	if ns < h.min {
		h.min = ns
	}
	if ns > h.max {
		h.max = ns
	}
	h.sum += ns
	idx := sort.Search(len(h.bounds), func(i int) bool {
		return d <= h.bounds[i]
	})
	h.counts[idx]++
}

func (h *latencyHistogram) merge(other *latencyHistogram) {
	if other == nil {
		return
	}
	if len(h.counts) != len(other.counts) {
		return
	}
	for i := range h.counts {
		h.counts[i] += other.counts[i]
	}
	h.sum += other.sum
	if other.min < h.min {
		h.min = other.min
	}
	if other.max > h.max {
		h.max = other.max
	}
}

func (h *latencyHistogram) quantile(q float64) time.Duration {
	if q <= 0 {
		if h.min == math.MaxInt64 {
			return 0
		}
		return time.Duration(h.min)
	}
	total := int64(0)
	for _, c := range h.counts {
		total += c
	}
	if total == 0 {
		return 0
	}
	target := int64(math.Ceil(float64(total) * q))
	acc := int64(0)
	for i, c := range h.counts {
		acc += c
		if acc >= target {
			if i >= len(h.bounds) {
				return h.bounds[len(h.bounds)-1]
			}
			return h.bounds[i]
		}
	}
	return h.bounds[len(h.bounds)-1]
}

func (h *latencyHistogram) mean() time.Duration {
	total := int64(0)
	for _, c := range h.counts {
		total += c
	}
	if total == 0 {
		return 0
	}
	return time.Duration(h.sum / total)
}

func formatDuration(d time.Duration) string {
	if int64(d) == math.MaxInt64 {
		return "0s"
	}
	return d.String()
}

// Metrics types.

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
