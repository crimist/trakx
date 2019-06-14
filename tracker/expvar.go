package tracker

import (
	"database/sql"
	"expvar"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func getUniqePeer() int64 {
	var num int64
	row := db.QueryRow("SELECT count(*) FROM peers")
	if err := row.Scan(&num); err != nil && err != sql.ErrNoRows {
		logger.Error("err", zap.Error(err))
	}
	return num
}

func getUniqeHash() int64 {
	var num int64
	row := db.QueryRow("SELECT count(distinct(hash)) FROM peers")
	if err := row.Scan(&num); err != nil && err != sql.ErrNoRows {
		logger.Error("err", zap.Error(err))
	}
	return num
}

func getUniqeIP() int64 {
	var num int64
	row := db.QueryRow("SELECT count(distinct(ip)) FROM peers")
	if err := row.Scan(&num); err != nil && err != sql.ErrNoRows {
		logger.Error("err", zap.Error(err))
	}
	return num
}

func getSeeds() int64 {
	var num int64
	row := db.QueryRow("SELECT count(*) FROM peers WHERE (complete = true)")
	if err := row.Scan(&num); err != nil && err != sql.ErrNoRows {
		logger.Error("err", zap.Error(err))
	}
	return num
}

func getLeeches() int64 {
	var num int64
	row := db.QueryRow("SELECT count(*) FROM peers WHERE (complete = false)")
	if err := row.Scan(&num); err != nil && err != sql.ErrNoRows {
		logger.Error("err", zap.Error(err))
	}
	return num
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
