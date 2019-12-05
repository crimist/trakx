package tracker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/crimist/trakx/bencoding"
	trakxhttp "github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/shared"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp"
	"github.com/honeybadger-io/honeybadger-go"
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
	var udptracker *udp.UDPTracker
	var err error

	// logger
	cfg := zap.NewDevelopmentConfig()
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	logger.Info("Starting trakx...")

	conf, err = shared.ViperConf(logger)
	if err != nil || !conf.Loaded() {
		logger.Panic("Failed to load configuration", zap.Any("config", conf), zap.Error(err))
	}

	logger.Info("dbg", zap.String("honey", conf.Trakx.Honey), zap.String("index", conf.Trakx.Index))

	if conf.Trakx.Honey != "" {
		logger.Info("Honeybadger.io API key detected", zap.String("keysample", conf.Trakx.Honey[:9]))
		honeybadger.Configure(honeybadger.Configuration{
			APIKey: conf.Trakx.Honey, // env TRAKX_TRACKER_HONEY
		})
		defer honeybadger.Monitor()
	}

	// db
	peerdb, err := storage.Open(conf)
	if err != nil {
		logger.Panic("Failed to initialize storage", zap.Error(err))
	}

	// pprof, sigs, expvar
	go sigHandler(peerdb, udptracker)
	if conf.Trakx.Pprof.Port != 0 {
		logger.Info("pprof enabled", zap.Int("port", conf.Trakx.Pprof.Port))
		initpprof()
	}

	// HTTP tracker / routes
	initRoutes()
	httptracker := trakxhttp.NewHTTPTracker(conf, logger, peerdb)

	if conf.Tracker.HTTP.Enabled {
		logger.Info("http tracker enabled", zap.Int("port", conf.Tracker.HTTP.Port))
		go httptracker.Serve(indexData, conf.Tracker.HTTP.Threads)
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
		udptracker = udp.NewUDPTracker(conf, logger, peerdb, conf.Tracker.UDP.Threads)
	}

	publishExpvar(conf, peerdb, httptracker, udptracker)
}
