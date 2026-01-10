package main

import (
	"math"
	"sort"
	"time"
)

const maxInt64 = int64(^uint64(0) >> 1)

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
		min:    maxInt64,
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
		if h.min == maxInt64 {
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
	if int64(d) == maxInt64 {
		return "0s"
	}
	return d.String()
}
