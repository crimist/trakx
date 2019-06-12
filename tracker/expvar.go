package tracker

import (
	"expvar"
	"go.uber.org/zap"
	"net/http"
)

func getUniqePeer() int64 {
	var peers []Peer
	err := db.Find(&peers).Error
	if err != nil {
		logger.Error("err", zap.Error(err))
	}
	return len(peers)
}

func getUniqeHash() int64 {
	var peers []Peer
	err := db.Select("DISTRINCT(hash)").Find(&peers).Error
	if err != nil {
		logger.Error("err", zap.Error(err))
	}
	return len(peers)
}

func getUniqeIP() int64 {
	var peers []Peer
	err := db.Select("DISTRINCT(ip)").Find(&peers).Error
	if err != nil {
		logger.Error("err", zap.Error(err))
	}
	return len(peers)
}

func getSeeds() int64 {
	var peers []Peer
	err := db.Where("complete = true").Find(&peers).Error
	if err != nil {
		logger.Error("err", zap.Error(err))
	}
	return len(peers)
}

func getLeeches() int64 {
	var peers []Peer
	err := db.Where("complete = false").Find(&peers).Error
	if err != nil {
		logger.Error("err", zap.Error(err))
	}
	return len(peers)
}

// Expvar is for netdata
func Expvar() {
	uniqueIP := expvar.NewInt("ip.unique")
	uniqueHash := expvar.NewInt("hash.unique")
	uniquePeer := expvar.NewInt("peer.unique")

	seeds := expvar.NewInt("tracker.seeds")
	leeches := expvar.NewInt("tracker.leeches")

	stats := expvar.NewMap("stats")
	stats.Set("seeds", new(expvar.Int))
	stats.Set("leeches", new(expvar.Int))

	go http.ListenAndServe("127.0.0.1:8080", nil) // only on localhost

	for {
		select {
		case <-tick.C:
			uniqueIP.Set(getUniqeIp())
			uniqueHash.Set(getUniqeHash())
			uniquePeer.Set(getUniqePeer())

			seeds.Set(getSeeds())
			leeches.Set(getLeeches())

			stats.Add("seeds", getSeeds())
			stats.Add("leeches", getLeeches())
		}
	}
}
