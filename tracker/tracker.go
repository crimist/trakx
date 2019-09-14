package tracker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/syc0x00/trakx/bencoding"
	trakxhttp "github.com/syc0x00/trakx/tracker/http"
	"github.com/syc0x00/trakx/tracker/shared"
	"github.com/syc0x00/trakx/tracker/storage"
	"github.com/syc0x00/trakx/tracker/udp"
	"go.uber.org/zap"

	// import database types so init is called
	_ "github.com/syc0x00/trakx/tracker/storage/map"
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

	conf = shared.ViperConf(logger)
	if conf.Tracker.Announce == 0 {
		logger.Fatal("Failed to load config")
		return
	}
	logger.Info("Loaded conf")

	// db
	peerdb, backup, err := storage.Open(conf.Database.Type, conf.Database.Backup)
	if err != nil {
		logger.Fatal("Failed to open database", zap.Error(err))
		return
	}
	peerdb.Init(conf, logger, backup)

	// pprof, sigs, expvar
	peerdb.Expvar()
	go handleSigs(peerdb, udptracker)
	if conf.Trakx.Pprof.Port != 0 {
		logger.Info("pprof on", zap.Int("port", conf.Trakx.Pprof.Port))
		initpprof()
	} else {
		logger.Info("pprof off")
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
