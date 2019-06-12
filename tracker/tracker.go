package tracker

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // gorm mysql
	"go.uber.org/zap"
)

const (
	trackerTimeout  = 60 * 45 // 45 min
	trackerInterval = 60 * 10 // 10 min
)

var (
	db     *gorm.DB
	logger *zap.Logger
	env    Enviroment
)

// Init initiates all the things the tracker needs
func Init(isProd bool) (*gorm.DB, error) {
	var err error
	var cfg zap.Config

	db, err = gorm.Open("mysql", "root@/trakx")
	if err != nil {
		return nil, err
	}

	if isProd == true {
		env = Prod
		cfg = zap.NewProductionConfig()
	} else {
		env = Dev
		cfg = zap.NewDevelopmentConfig()
		db.LogMode(true) // Debug gorm
	}

	cfg.OutputPaths = append(cfg.OutputPaths, "trakx.log")
	logger, err = cfg.Build()
	if err != nil {
		return nil, err
	}

	initPeer()
	initBans()

	return db, nil
}

// Clean removes clients that haven't checked in recently
func Clean() {
	for c := time.Tick(5 * time.Minute); ; <-c {
		affected := db.Where("last_seen < ?", time.Now().Unix()-int64(trackerTimeout)).Delete(&Peer{}).RowsAffected
		logger.Info("Cleaned peers", zap.Int64("count", affected))
		expvarCleaned += affected
	}
}
