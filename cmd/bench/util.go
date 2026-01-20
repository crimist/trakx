package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"math"
	"math/big"
	mrand "math/rand"
	"net/http"
	"time"
)

// Random number generation.

func cryptoSeed() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return time.Now().UnixNano()
	}
	return n.Int64()
}

func newRand(seed int64) *mrand.Rand {
	return mrand.New(mrand.NewSource(seed))
}

// Weighted choice for numwant distribution.

type weightedChoice struct {
	value      int
	cumulative float64
}

func chooseWeighted(rng *mrand.Rand, choices []weightedChoice, fallback int) int {
	if len(choices) == 0 {
		return fallback
	}
	r := rng.Float64()
	for _, c := range choices {
		if r <= c.cumulative {
			return c.value
		}
	}
	return choices[len(choices)-1].value
}

// Statistical distributions.

func clampInt(val, minVal, maxVal int) int {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

func drawLognormal(rng *mrand.Rand, median, sigma float64) int {
	if median <= 0 {
		median = 1
	}
	if sigma <= 0 {
		return int(math.Round(median))
	}
	mu := math.Log(median)
	return int(math.Round(math.Exp(mu + sigma*rng.NormFloat64())))
}

// pickNumwant returns a numwant value using the hardcoded weighted distribution.
func pickNumwant(rng *mrand.Rand) int {
	return chooseWeighted(rng, defaultNumwantWeights, 20)
}

// seedPeersForTorrent draws peer count using lognormal distribution with hardcoded realistic values.
func seedPeersForTorrent(rng *mrand.Rand) int {
	val := drawLognormal(rng, defaultSeedMedian, defaultSeedSigma)
	val = clampInt(val, defaultSeedMin, defaultSeedMax)
	if val < 1 {
		val = 1
	}
	return val
}

// Rate limiter.

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

// Stats fetching.

func fetchStats(url string, timeout time.Duration) (map[string]interface{}, error) {
	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
