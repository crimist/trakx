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

// Expvar is for netdata
func Expvar() {
	uniqueIP := expvar.NewInt("ip.unique")
	uniqueHash := expvar.NewInt("hash.unique")
	uniquePeer := expvar.NewInt("peer.unique")
	seeds := expvar.NewInt("tracker.seeds")
	leeches := expvar.NewInt("tracker.leeches")

	go http.ListenAndServe("127.0.0.1:1338", nil) // only on localhost

	tick := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-tick.C:
			uniqueIP.Set(getUniqeIP())
			uniqueHash.Set(getUniqeHash())
			uniquePeer.Set(getUniqePeer())
			seeds.Set(getSeeds())
			leeches.Set(getLeeches())
		}
	}
}
