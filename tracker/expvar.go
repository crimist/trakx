package tracker

import (
	"expvar"
	"net/http"
	"time"

	"github.com/thoas/stats"
)

func getInfo() (peers, hashes, ips, seeds, leeches int64) {
	ipmap := make(map[string]bool)

	for _, peermap := range PeerDB {
		peers += int64(len(peermap))

		for _, peer := range peermap {
			ipmap[peer.IP] = true
			if peer.Complete == true {
				seeds++
			} else {
				leeches++
			}
		}
	}

	hashes = int64(len(PeerDB))
	ips = int64(len(ipmap))

	return
}

var (
	expvarCleanedPeers  int64
	expvarCleanedHashes int64
	ExpvarAnnounces     int64
	ExpvarScrapes       int64
	ExpvarErrs          int64
)

// Expvar is for netdata
func Expvar(stats *stats.Stats) {
	var announcesOld int64
	var scrapesOld int64
	var errsOld int64

	uniqueIP := expvar.NewInt("tracker.ips")
	uniqueHash := expvar.NewInt("tracker.hashes")
	uniquePeer := expvar.NewInt("tracker.peers")

	cleanedPeers := expvar.NewInt("tracker.cleaned.peers")
	cleanedHashes := expvar.NewInt("tracker.cleaned.hashes")

	seeds := expvar.NewInt("tracker.seeds")
	leeches := expvar.NewInt("tracker.leeches")

	announces := expvar.NewInt("tracker.announces")
	announcesSec := expvar.NewInt("tracker.announces.sec")

	errors := expvar.NewInt("tracker.errors")
	errorsSec := expvar.NewInt("tracker.errors.sec")

	scrapes := expvar.NewInt("tracker.scrapes")
	scrapesSec := expvar.NewInt("tracker.scrapes.sec")

	respAvg := expvar.NewFloat("tracker.respavg") // milliseconds

	go http.ListenAndServe("127.0.0.1:"+trackerExpvarPort, nil) // only on localhost

	nextTime := time.Now().Truncate(time.Second)

	for {
		peers, hashes, ips, s, l := getInfo()
		uniqueIP.Set(ips)
		uniqueHash.Set(hashes)
		uniquePeer.Set(peers)

		seeds.Set(s)
		leeches.Set(l)

		cleanedPeers.Set(expvarCleanedPeers)
		cleanedHashes.Set(expvarCleanedHashes)

		announces.Set(ExpvarAnnounces)
		announcesSec.Set(ExpvarAnnounces - announcesOld)
		announcesOld = ExpvarAnnounces

		errors.Set(ExpvarErrs)
		errorsSec.Set(ExpvarErrs - errsOld)

		scrapes.Set(ExpvarScrapes)
		scrapesSec.Set(ExpvarScrapes - scrapesOld)
		scrapesOld = ExpvarScrapes

		respAvg.Set(stats.Data().AverageResponseTimeSec * 1000.0) // ms

		nextTime = nextTime.Add(time.Second)
		time.Sleep(time.Until(nextTime))
	}
}
