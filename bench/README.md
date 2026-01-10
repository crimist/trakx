# Trakx goroutine benchmark

This folder contains a load generator and orchestrators to benchmark the UDP/HTTP tracker
worker goroutines.
The goal is to find the smallest goroutine count that achieves near-peak throughput without
higher tail latency or error rates.

## Build

```bash
go build -o bench/trakxbench ./cmd/trakxbench
```

## Key assumptions

- Scrape ratio defaults to ~3% (roughly 20–50 scrapes per 1000 announces).
- UDP connection IDs are validated; the client performs a connect handshake and refreshes.
- Results reflect a typical Trakx deployment (no proxy headers, normal timeouts).

## Recommended server setup

For benchmarks, use the dedicated config in `bench/trakx.yaml`. It disables background
GC and periodic backups, and uses a separate cache path.

Config highlights:

- DB GC disabled (`db.gc: 0`)
- UDP connection GC disabled (`udp.connections.gc: 0`)
- Periodic backups disabled (`db.backup.interval: 0`)
- Maximums tracking disabled (`maximums.update_interval: 0`)

The orchestrator will launch Trakx with:

```bash
trakx --config bench/trakx.yaml start
```

It also sets `TRAKX_CACHE` to a benchmark-only directory and deletes it between runs.

## Orchestrated workflow (recommended)

You run one script on the server and one on the client. They coordinate automatically
so each goroutine count is started, benchmarked, stopped, and cleaned up.

Requirements: Python 3 and Go on both machines (the orchestrators build binaries every run).
Run these commands from the repo root so relative paths resolve.

### 1) On the server (runs Trakx)

```bash
bench/orchestrator_server.py --listen 0.0.0.0:9077 --config bench/trakx.yaml --cache-dir /tmp/trakx-bench-cache
```

Notes:

- The server listens for control messages on port 9077.
- It uses `trakx start` / `trakx stop` with the provided config.
- It clears the cache directory between runs (database reset).
- It sets `TRAKX_CACHE` so old backups are never reused.
- Do not point `--cache-dir` at any important data; it is deleted every run.
- The server script builds `trakx` every run (override path with `--trakx-bin`).

### 2) On the client (runs trakxbench)

```bash
bench/orchestrator_client.py --server 192.168.1.100:9077 --mode udp --goroutines 1,2,4,6,8,12,16,24,32,48,64 \
  --udp 192.168.1.100:1337 --http 192.168.1.100:1337 --stats-url http://192.168.1.100:1337/stats \
  --duration 2m --warmup 20s --concurrency 512 --scrape-ratio 0.03
```

The client will create `bench/results/<timestamp>-udp/` with one JSON file per run.

If you need extra `trakxbench` flags, pass them through with `--bench-args`.
The client script builds `trakxbench` every run (override path with `--bench-bin`).

### Manual step-through (optional)

If you want to press Enter between runs:

```bash
bench/orchestrator_server.py --manual ...
bench/orchestrator_client.py --pause ...
```

## Client flags overview

Common flags you will use:

- `--mode udp|http|both`: which tracker to benchmark.
- `--udp host:port` / `--http host:port`: tracker endpoints.
- `--duration` / `--warmup`: measurement and warmup times.
- `--concurrency`: number of client workers.
- `--rate`: target requests/sec (0 for max throughput).
- `--scrape-ratio`: fraction of requests that are scrapes.
- `--scrape-hashes`: info_hashes per scrape.
- `--stats`: expvar stats endpoint (optional).
- `--out`: JSON output path.

## Basic runs

UDP example (closed-loop, max throughput):

```bash
bench/trakxbench --mode udp --udp 192.168.1.100:1337 --duration 2m --warmup 20s --concurrency 512 --scrape-ratio 0.03 --stats http://192.168.1.100:1337/stats --out results-udp.json
```

HTTP example (open-loop, 1000 req/s target):

```bash
bench/trakxbench --mode http --http 192.168.1.100:1337 --rate 1000 --duration 2m --warmup 20s --concurrency 256 --scrape-ratio 0.03 --stats http://192.168.1.100:1337/stats --out results-http.json
```

## Seeding the DB (optional but recommended)

Seeding pre-populates the in-memory DB with a stable dataset so each run is comparable.

```bash
bench/trakxbench --mode udp --seed-torrents 10000 --seed-peers-per-torrent 25 --seed-only
```

Notes:

- Seeding sends announces with unique peer IDs.
- It does not simulate unique client IPs (Trakx uses remote IPs directly).
- Keep seed parameters the same across all goroutine runs.

## Testing a sweep of goroutine counts

You want to test a range of worker counts and compare throughput and tail latency.

Suggested sweep (adjust as needed): `1 2 4 6 8 12 16 24 32 48 64`

### Recommended

Use the orchestrated workflow (client + server scripts) so each run is clean, repeatable,
and fully automated. See the “Orchestrated workflow” section above.

## Best practices for accurate results

- Run the load generator on a different machine from the server when possible.
- Keep server CPU as the bottleneck (avoid client limits).
- Use a stable network path (avoid Wi-Fi, avoid background traffic).
- Use consistent parameters (scrape ratio, concurrency, duration).
- Use a warmup phase before measuring.
- Run each point at least 2–3 times and keep the median.
- Avoid mixing UDP and HTTP tests in the same run.
- Keep system power and CPU settings stable (avoid thermal throttling).
- If possible, pin `GOMAXPROCS` and record it for reproducibility.
 - Ensure the server cache directory is benchmark-only; the orchestrator deletes it per run.

## How to determine the optimal goroutine count

We want the smallest goroutine count that achieves near-peak throughput without harming
tail latency or increasing errors.

Recommended decision rule:

1) Find the peak throughput across all runs.
2) Select the smallest goroutine count that achieves at least 95% of peak throughput.
3) Ensure p95/p99 latency does not regress materially at that point.
4) Confirm error and timeout rates remain low and stable.

This avoids chasing marginal throughput gains at the cost of latency or stability.

## Result fields

The JSON output includes:

- Latency p50/p90/p95/p99.
- Success/error/timeout counts.
- Per-type counts (announce/scrape).
- Optional `/stats` snapshot.
- Runtime info (Go version, GOMAXPROCS).

Use these fields to compare runs and pick the knee of the curve.

## Notes on HTTP vs UDP

- UDP uses a connect handshake and must refresh connection IDs. The client handles this.
- HTTP is one request per TCP connection (server closes), so higher concurrency is often needed.
- Compare UDP and HTTP separately; their optimal goroutine counts may differ.
