package tracker

import (
	"expvar"
	"net/http"
	"time"
	"bytes"
)

func getUniqePeer() int64 {
	return int64(len(db))
}

func getUniqeHash() int64 {
	count := []Hash{}
	
	for _, val := range db {
		count = func() []Hash {
			for _, e := range count {
				if bytes.Equal(e, val.Hash) {
					return count
				}
			}
			return append(count, val.Hash)
		}()
	}

	return int64(len(count))
}

func getUniqeIP() int64 {
	count := []string{}
	
	for _, val := range db {
		count = func() []string {
			for _, e := range count {
				if e == val.IP {
					return count
				}
			}
			return append(count, val.IP)
		}()
	}

	return int64(len(count))
}

func getSeeds() int64 {
	var count int64
	for _, val := range db {
		if val.Complete == true {
			count++
		}
	}

	return count
}

func getLeeches() int64 {
	var count int64
	for _, val := range db {
		if val.Complete == false {
			count++
		}
	}

	return count
}

var (
	expvarCleaned int64
	expvarHits    int64
)

// Expvar is for netdata
func Expvar() {
	var oldHits int64

	uniqueIP := expvar.NewInt("tracker.ips")
	uniqueHash := expvar.NewInt("tracker.hashes")
	uniquePeer := expvar.NewInt("tracker.peers")
	cleaned := expvar.NewInt("tracker.cleaned")
	seeds := expvar.NewInt("tracker.seeds")
	leeches := expvar.NewInt("tracker.leeches")
	hits := expvar.NewInt("tracker.hits")
	hitsPerSec := expvar.NewInt("tracker.hitspersec")

	go http.ListenAndServe("127.0.0.1:1338", nil) // only on localhost

	for c := time.Tick(1 * time.Second); ; <-c {
		uniqueIP.Set(getUniqeIP())
		uniqueHash.Set(getUniqeHash())
		uniquePeer.Set(getUniqePeer())
		seeds.Set(getSeeds())
		leeches.Set(getLeeches())
		cleaned.Set(expvarCleaned)
		hits.Set(expvarHits)
		hitsPerSec.Set(expvarHits - oldHits)

		oldHits = expvarHits
	}
}
