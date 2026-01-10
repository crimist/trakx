#!/usr/bin/env python3
import argparse
import json
import os
import shlex
import socket
import subprocess
import sys
import time
from pathlib import Path
from typing import List, Tuple, Optional


def parse_hostport(value: str) -> Tuple[str, int]:
    if ":" not in value:
        raise ValueError("expected host:port")
    host, port_s = value.rsplit(":", 1)
    return host, int(port_s)


def parse_list(values: List[str]) -> List[int]:
    out = []
    for v in values:
        for part in v.split(","):
            part = part.strip()
            if not part:
                continue
            out.append(int(part))
    return out


def send_msg(fileobj, msg: dict) -> None:
    payload = json.dumps(msg).encode("utf-8") + b"\n"
    fileobj.write(payload)
    fileobj.flush()


def recv_msg(fileobj) -> Optional[dict]:
    line = fileobj.readline()
    if not line:
        return None
    return json.loads(line.decode("utf-8"))


def run_trakxbench(cmd: List[str]) -> int:
    proc = subprocess.run(cmd)
    return proc.returncode


def resolve_repo_root(value: Optional[str]) -> Path:
    if value:
        return Path(value).expanduser().resolve()
    return Path(__file__).resolve().parents[1]


def has_path_sep(value: str) -> bool:
    return os.sep in value or (os.altsep and os.altsep in value)


def ensure_trakxbench_binary(bin_arg: str, repo_root: Path) -> str:
    if not has_path_sep(bin_arg) and not bin_arg.startswith("."):
        bin_path = repo_root / bin_arg
    else:
        bin_path = Path(bin_arg)
        if not bin_path.is_absolute():
            bin_path = repo_root / bin_path

    bin_path.parent.mkdir(parents=True, exist_ok=True)
    build = subprocess.run(
        ["go", "build", "-o", str(bin_path), "./cmd/trakxbench"],
        cwd=str(repo_root),
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
    )
    if build.returncode != 0:
        raise RuntimeError(f"failed to build trakxbench: {build.stdout.strip()}")
    return str(bin_path)


def main() -> int:
    parser = argparse.ArgumentParser(description="Trakx benchmark client orchestrator")
    parser.add_argument("--server", default="127.0.0.1:9077", help="server control address")
    parser.add_argument("--mode", choices=["udp", "http"], required=True, help="benchmark mode")
    parser.add_argument("--goroutines", nargs="+", default=["1,2,4,6,8,12,16,24,32,48,64"], help="goroutine sweep list")
    parser.add_argument("--other-routines", type=int, default=1, help="routines for the non-tested tracker")
    parser.add_argument("--bench-bin", default="bench/bin/trakxbench", help="path to trakxbench binary (auto-built every run)")
    parser.add_argument("--out-dir", default="", help="output directory (default: bench/results/<timestamp>-<mode>)")
    parser.add_argument("--udp", default="", help="udp target host:port (default: <server>:1337)")
    parser.add_argument("--http", default="", help="http target host:port (default: <server>:1337)")
    parser.add_argument("--stats-url", default="", help="stats URL (default: http://<server>:1337/stats)")
    parser.add_argument("--duration", default="2m", help="benchmark duration")
    parser.add_argument("--warmup", default="20s", help="warmup duration")
    parser.add_argument("--concurrency", type=int, default=512, help="client workers")
    parser.add_argument("--rate", type=int, default=0, help="target requests/sec (0 max)")
    parser.add_argument("--scrape-ratio", type=float, default=0.03, help="scrape ratio")
    parser.add_argument("--scrape-hashes", type=int, default=5, help="hashes per scrape")
    parser.add_argument("--numwant", type=int, default=-1, help="numwant (-1 to omit)")
    parser.add_argument("--compact", action="store_true", help="use compact responses")
    parser.add_argument("--seed-torrents", type=int, default=0, help="seed torrents")
    parser.add_argument("--seed-peers-per-torrent", type=int, default=0, help="seed peers per torrent")
    parser.add_argument("--seed-fraction", type=float, default=0.2, help="fraction of peers that are seeds")
    parser.add_argument("--timeout", default="3s", help="per-request timeout")
    parser.add_argument("--udp-conn-refresh", default="10m", help="udp conn refresh interval")
    parser.add_argument("--label", default="", help="label prefix")
    parser.add_argument("--pause", action="store_true", help="pause for Enter between runs")
    parser.add_argument("--bench-args", default="", help="extra args passed to trakxbench")
    parser.add_argument("--repo-root", default="", help="repo root (default: inferred)")
    args = parser.parse_args()

    repo_root = resolve_repo_root(args.repo_root)
    args.bench_bin = ensure_trakxbench_binary(args.bench_bin, repo_root)

    goroutines = parse_list(args.goroutines)
    if not goroutines:
        print("no goroutine values provided", file=sys.stderr)
        return 2

    server_host, server_port = parse_hostport(args.server)
    if args.udp:
        udp_target = args.udp
    else:
        udp_target = f"{server_host}:1337"
    if args.http:
        http_target = args.http
    else:
        http_target = f"{server_host}:1337"
    if args.stats_url:
        stats_url = args.stats_url
    else:
        stats_url = f"http://{server_host}:1337/stats"

    if not args.out_dir:
        ts = time.strftime("%Y%m%d-%H%M%S")
        args.out_dir = str((repo_root / "bench" / "results" / f"{ts}-{args.mode}").resolve())
    os.makedirs(args.out_dir, exist_ok=True)

    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect((server_host, server_port))
    sock_file = sock.makefile("rwb")

    extra_args = shlex.split(args.bench_args) if args.bench_args else []

    for g in goroutines:
        if args.pause:
            input(f"Press Enter to start {args.mode} routines={g}...")

        if args.mode == "udp":
            udp_routines = g
            http_routines = args.other_routines
        else:
            udp_routines = args.other_routines
            http_routines = g

        send_msg(sock_file, {"type": "START", "udp_routines": udp_routines, "http_routines": http_routines})
        resp = recv_msg(sock_file)
        if not resp or resp.get("type") != "READY":
            print(f"server error: {resp}", file=sys.stderr)
            return 1

        out_path = os.path.join(args.out_dir, f"{args.mode}-{g}.json")
        label = args.label or args.mode
        if label:
            label = f"{label}-{g}"

        cmd = [
            args.bench_bin,
            "--mode", args.mode,
            "--udp", udp_target,
            "--http", http_target,
            "--stats", stats_url,
            "--duration", args.duration,
            "--warmup", args.warmup,
            "--concurrency", str(args.concurrency),
            "--rate", str(args.rate),
            "--scrape-ratio", str(args.scrape_ratio),
            "--scrape-hashes", str(args.scrape_hashes),
            "--numwant", str(args.numwant),
            "--timeout", args.timeout,
            "--udp-conn-refresh", args.udp_conn_refresh,
            "--out", out_path,
        ]
        if args.compact:
            cmd.append("--compact")
        if args.seed_torrents > 0 and args.seed_peers_per_torrent > 0:
            cmd.extend([
                "--seed-torrents", str(args.seed_torrents),
                "--seed-peers-per-torrent", str(args.seed_peers_per_torrent),
                "--seed-fraction", str(args.seed_fraction),
            ])
        if label:
            cmd.extend(["--label", label])
        cmd.extend(extra_args)

        print("running:", " ".join(cmd))
        code = run_trakxbench(cmd)

        send_msg(sock_file, {
            "type": "DONE",
            "udp_routines": udp_routines,
            "http_routines": http_routines,
            "result_path": out_path,
            "ok": code == 0,
        })
        resp = recv_msg(sock_file)
        if not resp or resp.get("type") != "STOPPED":
            print(f"server error: {resp}", file=sys.stderr)
            return 1

        if code != 0:
            print(f"trakxbench failed for {g}", file=sys.stderr)
            return code

    sock_file.close()
    sock.close()
    print(f"done. results in {args.out_dir}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
