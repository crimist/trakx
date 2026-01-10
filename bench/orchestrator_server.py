#!/usr/bin/env python3
import argparse
import json
import os
import shutil
import socket
import subprocess
import sys
import time
from typing import Optional, Tuple

HEARTBEAT_REQUEST = bytes([0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 0])
HEARTBEAT_OK = bytes([0xFF])


def parse_hostport(value: str) -> Tuple[str, int]:
    if ":" not in value:
        raise ValueError("expected host:port")
    host, port_s = value.rsplit(":", 1)
    return host, int(port_s)


def send_msg(fileobj, msg: dict) -> None:
    payload = json.dumps(msg).encode("utf-8") + b"\n"
    fileobj.write(payload)
    fileobj.flush()


def recv_msg(fileobj) -> Optional[dict]:
    line = fileobj.readline()
    if not line:
        return None
    return json.loads(line.decode("utf-8"))


def run_cmd(args, env=None) -> subprocess.CompletedProcess:
    return subprocess.run(args, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)


def wait_for_http(host: str, port: int, timeout: float) -> bool:
    import urllib.request

    url = f"http://{host}:{port}/heartbeat"
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            with urllib.request.urlopen(url, timeout=1) as resp:
                if resp.status == 200:
                    return True
        except Exception:
            time.sleep(0.2)
    return False


def wait_for_udp(host: str, port: int, timeout: float) -> bool:
    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            sock.settimeout(1)
            sock.sendto(HEARTBEAT_REQUEST, (host, port))
            data, _ = sock.recvfrom(16)
            if data == HEARTBEAT_OK:
                return True
        except Exception:
            time.sleep(0.2)
        finally:
            try:
                sock.close()
            except Exception:
                pass
    return False


def clear_cache(cache_dir: str) -> None:
    shutil.rmtree(cache_dir, ignore_errors=True)


def main() -> int:
    parser = argparse.ArgumentParser(description="Trakx benchmark server orchestrator")
    parser.add_argument("--listen", default="0.0.0.0:9077", help="control listen address")
    parser.add_argument("--trakx-bin", default="trakx", help="path to trakx binary")
    parser.add_argument("--config", default="bench/trakx.yaml", help="trakx config path")
    parser.add_argument("--cache-dir", default="/tmp/trakx-bench-cache", help="cache dir to reset between runs")
    parser.add_argument("--http-host", default="127.0.0.1", help="host for readiness check")
    parser.add_argument("--http-port", type=int, default=1337, help="http port for readiness check")
    parser.add_argument("--udp-host", default="127.0.0.1", help="host for udp heartbeat")
    parser.add_argument("--udp-port", type=int, default=1337, help="udp port for udp heartbeat")
    parser.add_argument("--ready-timeout", type=float, default=8.0, help="seconds to wait for readiness")
    parser.add_argument("--manual", action="store_true", help="pause for Enter before each start")
    args = parser.parse_args()

    listen_host, listen_port = parse_hostport(args.listen)
    env_base = os.environ.copy()
    env_base["TRAKX_CACHE"] = args.cache_dir

    server_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_sock.bind((listen_host, listen_port))
    server_sock.listen(1)
    print(f"listening on {listen_host}:{listen_port}")

    conn, addr = server_sock.accept()
    print(f"client connected from {addr[0]}:{addr[1]}")
    conn_file = conn.makefile("rwb")

    while True:
        msg = recv_msg(conn_file)
        if msg is None:
            print("client disconnected")
            break
        msg_type = msg.get("type")
        if msg_type == "START":
            if args.manual:
                input("Press Enter to start next run...")
            udp_routines = msg.get("udp_routines")
            http_routines = msg.get("http_routines")
            env = env_base.copy()
            if udp_routines is not None:
                env["TRAKX_UDP_ROUTINES"] = str(udp_routines)
            if http_routines is not None:
                env["TRAKX_HTTP_ROUTINES"] = str(http_routines)

            _ = run_cmd([args.trakx_bin, "--config", args.config, "stop"], env=env)
            clear_cache(args.cache_dir)

            start = run_cmd([args.trakx_bin, "--config", args.config, "start"], env=env)
            if start.returncode != 0:
                send_msg(conn_file, {"type": "ERROR", "message": start.stdout.strip()})
                continue

            ready = False
            if args.http_port > 0:
                ready = wait_for_http(args.http_host, args.http_port, args.ready_timeout)
            if not ready and args.udp_port > 0:
                ready = wait_for_udp(args.udp_host, args.udp_port, args.ready_timeout)
            if not ready:
                send_msg(conn_file, {"type": "ERROR", "message": "tracker not ready"})
                continue

            send_msg(conn_file, {"type": "READY", "udp_routines": udp_routines, "http_routines": http_routines})
        elif msg_type == "DONE":
            udp_routines = msg.get("udp_routines")
            http_routines = msg.get("http_routines")
            env = env_base.copy()
            if udp_routines is not None:
                env["TRAKX_UDP_ROUTINES"] = str(udp_routines)
            if http_routines is not None:
                env["TRAKX_HTTP_ROUTINES"] = str(http_routines)

            stop = run_cmd([args.trakx_bin, "--config", args.config, "stop"], env=env)
            if stop.returncode != 0:
                send_msg(conn_file, {"type": "ERROR", "message": stop.stdout.strip()})
                continue
            send_msg(conn_file, {"type": "STOPPED", "udp_routines": udp_routines, "http_routines": http_routines})
        elif msg_type == "PING":
            send_msg(conn_file, {"type": "PONG"})
        else:
            send_msg(conn_file, {"type": "ERROR", "message": f"unknown message type: {msg_type}"})

    conn_file.close()
    conn.close()
    server_sock.close()
    return 0


if __name__ == "__main__":
    sys.exit(main())
