package tracker

import (
	"expvar"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func getUniqePeer() int64 {
	var num uint
	if err := db.Model(&Peer{}).Count(&num).Error; err != nil {
		logger.Error("err", zap.Error(err))
	}
	return int64(num)
}

func getUniqeHash() int64 {
	var num uint
	if err := db.Model(&Peer{}).Select("count(distinct(peers.hash))").Count(&num).Error; err != nil {
		logger.Error("err", zap.Error(err))
	}
	return int64(num)
}

func getUniqeIP() int64 {
	var num uint
	if err := db.Model(&Peer{}).Select("count(distinct(peers.ip))").Count(&num).Error; err != nil {
		logger.Error("err", zap.Error(err))
	}
	return int64(num)
}

func getSeeds() int64 {
	var num uint
	if err := db.Model(&Peer{}).Where("complete = true").Count(&num).Error; err != nil {
		logger.Error("err", zap.Error(err))
	}
	return int64(num)
}

func getLeeches() int64 {
	var num uint
	if err := db.Model(&Peer{}).Where("complete = false").Count(&num).Error; err != nil {
		logger.Error("err", zap.Error(err))
	}
	return int64(num)
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

	for c := time.Tick(1 * time.Minute); ; <-c {
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
