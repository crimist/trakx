package tracker

import (
	"time"

	"go.uber.org/zap"
)

const (
	trackerTimeout       = 60 * 45 // 45 min
	trackerInterval      = 60 * 15 // 15 min
	trackerCleanInterval = 3 * time.Minute
)

type ID [20]byte

var (
	db     map[ID]Peer
	logger *zap.Logger
	env    Enviroment
)

// Init initiates all the things the tracker needs
func Init(isProd bool) (error) {
	var err error
	var cfg zap.Config
	db = make(map[ID]Peer)

	if isProd == true {
		env = Prod
		cfg = zap.NewProductionConfig()
	} else {
		env = Dev
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.OutputPaths = append(cfg.OutputPaths, "trakx.log")
	logger, err = cfg.Build()
	
	return err
}

// Clean removes clients that haven't checked in recently
func Clean() {
	for c := time.Tick(trackerCleanInterval); ; <-c {
		for key, val := range db {
			if val.LastSeen < time.Now().Unix()-int64(trackerTimeout) {
				delete(db, key)
				expvarCleaned++
			}
		}
	}
}
