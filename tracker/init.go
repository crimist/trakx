package tracker

import (
	"time"

	"github.com/jinzhu/gorm"
	"go.uber.org/zap"

	// gorm mysql
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db     *gorm.DB
	logger *zap.Logger
	env    Enviroment
)

// Init initiates all the things the tracker needs
func Init(isProd bool) (*gorm.DB, error) {
	var err error

	if isProd == true {
		env = Prod
		logger, err = zap.NewProduction()
	} else {
		env = Dev
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return nil, err
	}

	db, err = gorm.Open("mysql", "root@/bittorrent")
	if err != nil {
		return nil, err
	}

	initPeer()
	initBans()

	return db, nil
}

// Clean removes clients that havn't checked in recently
func Clean() {
	for c := time.Tick(5 * time.Minute); ; <-c {
		// Delete if hasn't checked in in 30 min
		affected := db.Where("last_seen < ?", time.Now().Unix()-int64(60*30)).Delete(&Peer{}).RowsAffected
		logger.Info("Cleaned peers",
			zap.Int64("count", affected),
		)
	}
}
