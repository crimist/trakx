package tracker

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/crimist/trakx/bencoding"
	trakxhttp "github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	// import database types so init is called
	_ "github.com/crimist/trakx/tracker/storage/map"
)

var (
	logger *zap.Logger
	conf   *shared.Config
)

// Run runs the tracker
func Run() {
	var udptracker udp.UDPTracker
	var httptracker trakxhttp.HTTPTracker
	var err error

	rand.Seed(time.Now().UnixNano() * time.Now().Unix())

	cfg := zap.NewDevelopmentConfig()
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	logger.Info("Starting trakx...")

	conf, err = shared.LoadConf(logger)
	if err != nil {
		logger.Warn("Failed to load a configuration", zap.Any("config", conf), zap.Error(errors.WithMessage(err, "Failed to load viper cofig")))
	}
	if !conf.Loaded() {
		logger.Fatal("Config failed to load critical values")
	}

	// db
	peerdb, err := storage.Open(conf)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	// pprof, sigs, expvar
	go sigHandler(peerdb, &udptracker, &httptracker)
	if conf.Trakx.Pprof.Port != 0 {
		logger.Info("pprof enabled", zap.Int("port", conf.Trakx.Pprof.Port))
		initpprof()
	}

	// routes
	initRoutes()

	if conf.Tracker.HTTP.Enabled {
		logger.Info("http tracker enabled", zap.Int("port", conf.Tracker.HTTP.Port))

		httptracker.Init(conf, logger, peerdb)
		go httptracker.Serve(string(indexData)) // indexData in the routes.go file
	} else {
		d := bencoding.NewDict()
		d.Int64("interval", 432000) // 5 days
		errResp := []byte(d.Get())

		trackerMux := http.NewServeMux()
		trackerMux.HandleFunc("/", index)
		trackerMux.HandleFunc("/dmca", dmca)
		trackerMux.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {})
		trackerMux.HandleFunc("/announce", func(w http.ResponseWriter, r *http.Request) {
			w.Write(errResp)
		})

		server := http.Server{
			Addr:         fmt.Sprintf(":%d", conf.Tracker.HTTP.Port),
			Handler:      trackerMux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 7 * time.Second,
			IdleTimeout:  0,
		}
		server.SetKeepAlivesEnabled(false)

		go func() {
			if err := server.ListenAndServe(); err != nil {
				logger.Error("ListenAndServe()", zap.Error(err))
			}
		}()
	}

	// UDP tracker
	if conf.Tracker.UDP.Enabled {
		logger.Info("udp tracker enabled", zap.Int("port", conf.Tracker.UDP.Port))
		udptracker.Init(conf, logger, peerdb)
		go udptracker.Serve()
	}

	publishExpvar(conf, peerdb, &httptracker, &udptracker)
}
