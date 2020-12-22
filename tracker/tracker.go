package tracker

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
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
	conf *shared.Config
)

// Run runs the tracker
func Run() {
	var udptracker udp.UDPTracker
	var httptracker trakxhttp.HTTPTracker
	var err error

	rand.Seed(time.Now().UnixNano() * time.Now().Unix())

	conf, err = shared.LoadConf()
	if err != nil {
		if conf.Logger != nil {
			conf.Logger.Warn("Failed to load a configuration", zap.Any("config", conf), zap.Error(errors.WithMessage(err, "Failed to load config")))
		}
		println("failed to load config and logger", err, conf)
	}
	if !conf.Loaded() {
		println("Config failed to load critical values")
		os.Exit(1)
	}

	shared.LoadEmbed(conf.Logger)

	conf.Logger.Info("Loaded conf and embed, starting trakx...")

	// db
	peerdb, err := storage.Open(conf)
	if err != nil {
		conf.Logger.Fatal("Failed to initialize storage", zap.Error(err))
	}

	// init the peerchan q with minimum
	storage.PeerChan.Add(conf.PeerChanMin)

	// pprof, sigs, expvar
	go sigHandler(peerdb, &udptracker, &httptracker)
	if conf.PprofPort != 0 {
		conf.Logger.Info("pprof enabled", zap.Int("port", conf.PprofPort))
		initpprof()
	}

	if conf.Tracker.HTTP.Enabled {
		conf.Logger.Info("http tracker enabled", zap.Int("port", conf.Tracker.HTTP.Port))

		httptracker.Init(conf, peerdb)
		go httptracker.Serve()
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
				conf.Logger.Error("ListenAndServe()", zap.Error(err))
			}
		}()
	}

	// UDP tracker
	if conf.Tracker.UDP.Enabled {
		conf.Logger.Info("udp tracker enabled", zap.Int("port", conf.Tracker.UDP.Port))
		udptracker.Init(conf, peerdb)
		go udptracker.Serve()
	}

	publishExpvar(conf, peerdb, &httptracker, &udptracker)
}
