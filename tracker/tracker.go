package tracker

import (
	"time"

	_ "github.com/go-sql-driver/mysql" // mysql driver
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const (
	trackerTimeout       = 60 * 45 // 45 min
	trackerInterval      = 60 * 15 // 15 min
	trackerCleanInterval = 3 * time.Minute
)

var (
	db     *sqlx.DB
	logger *zap.Logger
	env    Enviroment
)

// Init initiates all the things the tracker needs
func Init(isProd bool) (*sqlx.DB, error) {
	var err error
	var cfg zap.Config

	db, err = sqlx.Open("mysql", "root@/trakx")
	if err != nil {
		return nil, err
	}

	if isProd == true {
		env = Prod
		cfg = zap.NewProductionConfig()
	} else {
		env = Dev
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.OutputPaths = append(cfg.OutputPaths, "trakx.log")
	logger, err = cfg.Build()
	if err != nil {
		return nil, err
	}

	initPeerTable()
	initBans()

	return db, nil
}

// Clean removes clients that haven't checked in recently
func Clean() {
	for c := time.Tick(trackerCleanInterval); ; <-c {
		result, err := db.Exec("DELETE from `peers` WHERE (last_seen < ?)", time.Now().Unix()-int64(trackerTimeout))
		if err != nil {
			logger.Error("Peer delete", zap.Error(err))
		}
		affected, err := result.RowsAffected()
		if err != nil {
			logger.Error("Affected", zap.Error(err))
		}
		logger.Info("Cleaned peers", zap.Int64("count", affected))
		expvarCleaned += affected
	}
}
