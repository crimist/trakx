package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var client = &http.Client{}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()-_=+[{]}\\|\":;<,>.?'/~`"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

func randStr(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

func reqFast(infoHash, ip, event, left, peerID, key, port string, compact bool) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:1337/announce", nil)
	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	q.Add("info_hash", infoHash)
	q.Add("ip", ip)
	q.Add("event", event)
	q.Add("left", left)
	q.Add("peer_id", peerID)
	q.Add("key", key)
	q.Add("port", port)
	if compact {
		q.Add("compact", "1")
	}
	req.URL.RawQuery = q.Encode()

	_, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	rand.Seed(time.Now().Unix())

	for i := 0; i < 50; i++ {
		fmt.Println("i", i)
		hash := randStr(20)
		id := randStr(20)
		reqFast(hash, "69.69.69.69", "started", "1000", id, "key", "42069", false)
		if rand.Int63()&(1<<62) == 0 {
			reqFast(hash, "69.69.69.69", "complete", "1000", id, "key", "42069", false)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
