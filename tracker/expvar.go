package tracker

import (
	"expvar"
	"net/http"
	"time"
)

func getInfo() (int64, int64, int64, int64, int64) {
	var peers int64
	var seeds int64
	var leeches int64
	ipmap := make(map[string]bool)

	for _, peermap := range db {
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

	hashes := int64(len(db))
	ips := int64(len(ipmap))

	return peers, hashes, ips, seeds, leeches
}

var (
	expvarCleanedPeers  int64
	expvarCleanedHashes int64
	expvarAnnounces     int64
	expvarScrapes       int64
	expvarErrs          int64
)

// Expvar is for netdata
func Expvar() {
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

		announces.Set(expvarAnnounces)
		announcesSec.Set(expvarAnnounces - announcesOld)
		announcesOld = expvarAnnounces

		errors.Set(expvarErrs)
		errorsSec.Set(expvarErrs - errsOld)

		scrapes.Set(expvarScrapes)
		scrapesSec.Set(expvarScrapes - scrapesOld)
		scrapesOld = expvarScrapes

		nextTime = nextTime.Add(time.Second)
		time.Sleep(time.Until(nextTime))
	}
}
