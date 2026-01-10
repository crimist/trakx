package main

import (
	"fmt"
	"math"
	mrand "math/rand"
	"strconv"
	"strings"
)

type weightedChoice struct {
	value      int
	weight     float64
	cumulative float64
}

func parseWeightedChoices(spec string) ([]weightedChoice, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("numwant-weights is empty")
	}
	parts := strings.Split(spec, ",")
	var choices []weightedChoice
	total := 0.0
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.Split(part, ":")
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid numwant-weights entry: %q", part)
		}
		val, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			return nil, fmt.Errorf("invalid numwant value: %q", fields[0])
		}
		weight, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid numwant weight: %q", fields[1])
		}
		if weight <= 0 {
			return nil, fmt.Errorf("numwant weight must be > 0")
		}
		choices = append(choices, weightedChoice{value: val, weight: weight})
		total += weight
	}
	if len(choices) == 0 {
		return nil, fmt.Errorf("no valid numwant weights")
	}
	if total <= 0 {
		return nil, fmt.Errorf("invalid total numwant weight")
	}
	cumulative := 0.0
	for i := range choices {
		cumulative += choices[i].weight / total
		choices[i].cumulative = cumulative
	}
	choices[len(choices)-1].cumulative = 1.0
	return choices, nil
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

func clampInt(val, minVal, maxVal int) int {
	if minVal > 0 && val < minVal {
		return minVal
	}
	if maxVal > 0 && val > maxVal {
		return maxVal
	}
	return val
}

func drawNormal(rng *mrand.Rand, mean, sigma float64) int {
	if sigma <= 0 {
		return int(math.Round(mean))
	}
	return int(math.Round(mean + sigma*rng.NormFloat64()))
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

func (cfg *config) pickNumwant(rng *mrand.Rand) int {
	if cfg.numwantDist == "weighted" {
		return chooseWeighted(rng, cfg.numwantChoices, cfg.numwant)
	}
	return cfg.numwant
}

func (cfg *config) seedPeersForTorrent(rng *mrand.Rand) (int, error) {
	var val int
	switch cfg.seedPeersDist {
	case "fixed":
		val = cfg.seedPeersPerTor
	case "normal":
		val = drawNormal(rng, cfg.seedPeersMean, cfg.seedPeersSigma)
	case "lognormal":
		val = drawLognormal(rng, cfg.seedPeersMedian, cfg.seedPeersSigma)
	default:
		return 0, fmt.Errorf("invalid seed-peers-dist: %s", cfg.seedPeersDist)
	}

	val = clampInt(val, cfg.seedPeersMin, cfg.seedPeersMax)
	if val < 1 {
		val = 1
	}
	return val, nil
}
