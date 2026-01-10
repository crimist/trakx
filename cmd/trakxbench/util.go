package main

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand"
	"time"
)

func cryptoSeed() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(maxInt64))
	if err != nil {
		return time.Now().UnixNano()
	}
	return n.Int64()
}

func newRand(seed int64) *mrand.Rand {
	return mrand.New(mrand.NewSource(seed))
}
