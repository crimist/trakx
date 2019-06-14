package tracker

import (
	"expvar"
	"net/http"
	"time"
)

func getInfo() (peers, hashes, ips, seeds, leeches int64) {
	ipArr := []string{}

	for _, peermap := range db {
		peers += int64(len(peermap))

		for _, peer := range peermap {
			ipArr = func() []string {
				for _, ip := range ipArr {
					if ip == peer.IP {
						return ipArr
					}
				}
				return append(ipArr, peer.IP)
			}()

			if peer.Complete == true {
				seeds++
			} else {
				leeches++
			}
		}
	}

	hashes = int64(len(db))
	ips = int64(len(ipArr))

	return peers, hashes, ips, seeds, leeches
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
		peers, hashes, ips, s, l := getInfo()
		uniqueIP.Set(ips)
		uniqueHash.Set(hashes)
		uniquePeer.Set(peers)
		seeds.Set(s)
		leeches.Set(l)
		cleaned.Set(expvarCleaned)
		hits.Set(expvarHits)
		hitsPerSec.Set(expvarHits - oldHits)

		oldHits = expvarHits
	}
}
