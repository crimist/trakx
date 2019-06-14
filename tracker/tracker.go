package tracker

import (
	"time"

	"go.uber.org/zap"
)

const (
	trackerTimeout       = 60 * 45 // 45 min
	trackerInterval      = 60 * 20 // 15 min
	trackerCleanInterval = 3 * time.Minute
)

var (
	db     map[Hash]map[PeerID]Peer
	logger *zap.Logger
	env    Enviroment
)

// Init initiates all the things the tracker needs
func Init(isProd bool) error {
	var err error
	var cfg zap.Config
	db = make(map[Hash]map[PeerID]Peer)

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
		for hash, peermap := range db {
			if len(peermap) == 0 {
				delete(db, hash)
				continue
			}
			for id, peer := range peermap {
				if peer.LastSeen < time.Now().Unix()-int64(trackerTimeout) {
					delete(peermap, id)
					expvarCleaned++
				}
			}
		}
	}
}
