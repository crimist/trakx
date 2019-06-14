package tracker

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

const (
	trackerCleanTimeout     = 45 * time.Minute
	trackerAnnounceInterval = 20 * time.Minute
	trackerCleanInterval    = 3 * time.Minute
	trackerWriteDBInterval  = 5 * time.Minute
	trackerDBFilename       = "trakx.db"
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

	loadDB()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-c
		logger.Info("Got signal", zap.Any("Signal", sig))

		writeDB()

		os.Exit(128 + int(sig.(syscall.Signal)))
	}()

	go func() {
		for c := time.Tick(trackerWriteDBInterval); ; <-c {
			writeDB()
		}
	}()

	return err
}

func loadDB() {
	file, err := os.Open(trackerDBFilename)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("No database found")
			return
		}
		logger.Panic("db open", zap.Error(err))
	}
	decoder := gob.NewDecoder(file)

	if err := decoder.Decode(&db); err != nil {
		logger.Panic("db gob decoder", zap.Error(err))
	}

	logger.Info("Loaded database", zap.Int("hashes", len(db)))
}

// Write dumps the database to a file
func writeDB() {
	buff := new(bytes.Buffer)
	encoder := gob.NewEncoder(buff)

	if err := encoder.Encode(db); err != nil {
		logger.Panic("db gob encoder", zap.Error(err))
	}

	if err := ioutil.WriteFile(trackerDBFilename, buff.Bytes(), 0644); err != nil {
		logger.Panic("db writefile", zap.Error(err))
	}

	logger.Info("Wrote database", zap.Int("hashes", len(db)))
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
				if peer.LastSeen < time.Now().Unix()-int64(trackerCleanTimeout) {
					delete(peermap, id)
					expvarCleaned++
				}
			}
		}
	}
}
