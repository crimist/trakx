import sys
import requests
import random
import time
import threading

# curl -sK -v http://localhost:1338/debug/pprof/heap > heap.out; pprof -http=127.0.0.1:6006 heap.out
url = "http://127.0.0.1:1337/announce"

def stress():
    while True:
        hash = bytearray(random.getrandbits(8) for _ in xrange(20))

        # add 10 peers to hash
        for _ in range(10):
            id = bytearray(random.getrandbits(8) for _ in xrange(20))
            payload = {"info_hash": str(hash), "peer_id": str(id), "port": "1"}
            requests.get(url, params=payload)
            sys.stdout.write(".")

for _ in range(10):
    t = threading.Thread(target=stress)
    t.daemon = True
    t.start()

while True:
    sys.stdout.flush()
    time.sleep(1)
