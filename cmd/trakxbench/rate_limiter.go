package main

import (
	"context"
	"time"
)

type rateLimiter struct {
	tokens <-chan struct{}
}

func newRateLimiter(ctx context.Context, rate int) *rateLimiter {
	if rate <= 0 {
		return nil
	}
	interval := time.Duration(float64(time.Second) / float64(rate))
	if interval <= 0 {
		interval = time.Nanosecond
	}
	ch := make(chan struct{}, rate)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}()
	return &rateLimiter{tokens: ch}
}
