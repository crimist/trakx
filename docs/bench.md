# Trakx Benchmark Tools

Tools for benchmarking the trakx UDP/HTTP tracker:

- **bench** - Load generator that sends announce/scrape requests
- **orchestrate** - Orchestrator for automated benchmark sweeps (local and distributed)

## Quick Start

### Build Tools

```bash
# Build both tools
go build -o bench/bin/bench ./cmd/bench
go build -o bench/bin/orchestrate ./cmd/orchestrate
```

### Run a Local Benchmark Sweep

```bash
# Simple goroutine sweep (builds binaries automatically)
./bench/bin/orchestrate run --routines 4,8,16,32 --mode udp

# With custom scenario file
./bench/bin/orchestrate run --scenario configs/bench/scenarios/comprehensive.yaml
```

### Run a Distributed Benchmark

```bash
# On server machine (192.168.1.100) - manages trakx tracker
./bench/bin/orchestrate server --listen 0.0.0.0:9077

# On client machine - runs benchmarks against server
./bench/bin/orchestrate client --server 192.168.1.100:9077 --routines 4,8,16,32
```

---

## bench - Load Generator

A simple, focused load generator for trakx.

### Flags (10 total)

| Flag | Default | Purpose |
|------|---------|---------|
| `-mode` | `udp` | Protocol: `udp`, `http`, or `both` |
| `-target` | `127.0.0.1:1337` | Tracker address (host:port) |
| `-duration` | `2m` | Measurement duration |
| `-rate` | `0` | Requests/sec (0 = max throughput) |
| `-concurrency` | `64` | Worker count |
| `-seed` | `0` | Pre-seed with N torrents (0 = disabled) |
| `-rng-seed` | `0` | RNG seed (0 = random) |
| `-out` | stdout | Output file path |
| `-stats` | `""` | Expvar stats URL (optional) |
| `-quiet` | `false` | Suppress progress output |

### Hardcoded Realistic Defaults

All BitTorrent workload parameters are baked in:

- Warmup: 10s, timeout: 3s, connection refresh: 10min
- Scrape ratio: 3%, scrape hashes: 5, torrents per peer: 4
- Numwant distribution: 20 (90%), 30 (8%), 50 (2%)
- Seed peer distribution: lognormal (median=30, sigma=1.2)
- Seeder fraction: 20%

### Example Usage

```bash
# Simple benchmark
./bench -mode udp -seed 2048 -out results.json

# Rate-limited benchmark
./bench -mode udp -rate 10000 -seed 2048 -out results.json

# With stats collection
./bench -mode udp -seed 2048 -stats http://127.0.0.1:1337/stats -out results.json
```

---

## orchestrate - Benchmark Orchestrator

Automates benchmark sweeps with tracker lifecycle management.

### Commands

- `run` - Local mode: runs everything on one machine
- `server` - Distributed mode: runs on tracker machine
- `client` - Distributed mode: runs on benchmark machine

### Local Mode (`run`)

Manages the complete lifecycle on a single machine:
1. Builds binaries
2. For each goroutine count:
   - Stops any running tracker
   - Clears cache
   - Starts tracker with configured routines
   - Runs benchmark
   - Collects results
3. Analyzes and summarizes results

```bash
orchestrate run [options]
  --scenario <file>     Scenario YAML file
  --routines <list>     Goroutine counts (default: 4,8,16,32)
  --mode <mode>         Protocol: udp, http (default: udp)
  --duration <dur>      Benchmark duration (default: 2m)
  --seed <n>            Seed torrents (default: 2048)
  --output <dir>        Output directory
  --config <file>       Trakx config (default: configs/bench/trakx.yaml)
  --pause               Pause between runs
  --quiet               Suppress progress
```

### Distributed Mode

For realistic network testing with separate machines.

**Server** (runs on tracker machine):
```bash
orchestrate server [options]
  --listen <addr>       Listen address (default: 0.0.0.0:9077)
  --config <file>       Trakx config (default: configs/bench/trakx.yaml)
  --http-addr <addr>    HTTP readiness address (default: 127.0.0.1:1337)
  --udp-addr <addr>     UDP readiness address (default: 127.0.0.1:1337)
  --ready-timeout <dur> Readiness timeout (default: 30s)
```

**Client** (runs on benchmark machine):
```bash
orchestrate client [options]
  --server <addr>       Server address (required)
  --scenario <file>     Scenario YAML file
  --routines <list>     Goroutine counts (default: 4,8,16,32)
  --mode <mode>         Protocol: udp, http (default: udp)
  --duration <dur>      Benchmark duration (default: 2m)
  --seed <n>            Seed torrents (default: 2048)
  --target <addr>       Tracker address (default: server:1337)
  --output <dir>        Output directory
  --pause               Pause between runs
```

### Protocol

The client and server communicate via JSON messages over TCP:

- `START` → `READY` - Start tracker with goroutine config
- `DONE` → `STOPPED` - Benchmark complete, stop tracker
- `PING` → `PONG` - Heartbeat

---

## Scenario Files

YAML files that define benchmark configurations.

### Structure

```yaml
name: my-scenario
description: Description of what this tests

defaults:
  duration: 2m
  concurrency: 64
  seed: 2048
  rate: 0  # 0 = max throughput

runs:
  - name: udp-sweep
    mode: udp
    routines: [1, 2, 4, 8, 16, 32]
    other_routines: 1  # routines for the other protocol
    overrides:         # optional per-run overrides
      duration: 1m
      concurrency: 512
```

### Built-in Scenarios

- `configs/bench/scenarios/default.yaml` - Standard UDP sweep
- `configs/bench/scenarios/comprehensive.yaml` - UDP, HTTP, and rate-limited

---

## Output Format

### Per-Run Results (bench)

```json
{
  "generated_at": "2024-01-15T10:30:00Z",
  "mode": "udp",
  "target": "127.0.0.1:1337",
  "config": { ... },
  "metrics": {
    "udp": {
      "count": 1234567,
      "success": 1234560,
      "errors": 5,
      "timeouts": 2,
      "rate_per_sec": 10288.05,
      "latency": {
        "p50": "102.4µs",
        "p90": "204.8µs",
        "p95": "409.6µs",
        "p99": "819.2µs"
      }
    }
  }
}
```

### Orchestration Summary (orchestrate)

```json
{
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "scenario": "default",
  "runs": [
    {
      "run_id": "udp-sweep-4",
      "routines": 4,
      "status": "ok",
      "summary": {
        "throughput": 45231,
        "p99": "1.6ms"
      }
    }
  ],
  "analysis": {
    "peak_throughput": 125678,
    "peak_config": { "routines": 32 },
    "optimal_config": { "routines": 16 },
    "optimal_reason": "Smallest config achieving >= 95% of peak"
  }
}
```

---

## Best Practices

1. **Separate machines**: Run client on different machine for realistic results
2. **Stable network**: Avoid Wi-Fi and background traffic
3. **Consistent parameters**: Use same seed, duration, concurrency
4. **Multiple runs**: Run each config 2-3 times, take median
5. **Stable CPU**: Avoid thermal throttling, pin GOMAXPROCS
6. **Clean state**: Use `configs/bench/trakx.yaml` which disables GC and backups

---

## Determining Optimal Goroutine Count

The orchestrator automatically analyzes results:

1. Finds peak throughput across all runs
2. Selects smallest goroutine count achieving >= 95% of peak
3. Reports both peak and optimal configurations

Example output:
```
Peak:    32 routines @ 125678 req/s (p99: 2.8ms)
Optimal: 16 routines @ 112845 req/s (p99: 2.1ms)
         Smallest config achieving >= 95% of peak (119394 req/s)
```
