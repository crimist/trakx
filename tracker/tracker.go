package tracker

import (
	"fmt"
	"math/rand"
	gohttp "net/http"
	"time"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/storage"
	"github.com/crimist/trakx/tracker/udp"
	"go.uber.org/zap"

	// import database types so init is called
	_ "github.com/crimist/trakx/tracker/storage/map"
)

// Run runs the tracker
func Run() {
	var udptracker udp.UDPTracker
	var httptracker http.HTTPTracker
	var err error

	rand.Seed(time.Now().UnixNano() * time.Now().Unix())

	if !config.Conf.Loaded() {
		config.Logger.Fatal("Config failed to load critical values")
	}

	config.Logger.Info("Loaded configuration, starting trakx...")

	// db
	peerdb, err := storage.Open()
	if err != nil {
		config.Logger.Fatal("Failed to initialize storage", zap.Error(err))
	} else {
		config.Logger.Info("Initialized storage")
	}

	// init the peerchan with minimum
	storage.PeerChan.Add(config.Conf.Debug.PeerChanInit)

	// run signal handler
	go signalHandler(peerdb, &udptracker, &httptracker)

	// init pprof if enabled
	if config.Conf.Debug.PprofPort != 0 {
		config.Logger.Info("pprof enabled", zap.Int("port", config.Conf.Debug.PprofPort))
		initpprof()
	}

	if config.Conf.Tracker.HTTP.Mode == config.TrackerModeEnabled {
		config.Logger.Info("http tracker enabled", zap.Int("port", config.Conf.Tracker.HTTP.Port))

		httptracker.Init(peerdb)
		go httptracker.Serve()
	} else if config.Conf.Tracker.HTTP.Mode == config.TrackerModeInfo {
		// serve basic html server with index and dmca pages
		d := bencoding.NewDictionary()
		d.Int64("interval", 86400) // 1 day
		errResp := d.GetBytes()

		mux := gohttp.NewServeMux()
		mux.HandleFunc("/", index)
		mux.HandleFunc("/dmca", dmca)
		mux.HandleFunc("/scrape", func(w gohttp.ResponseWriter, r *gohttp.Request) {})
		mux.HandleFunc("/announce", func(w gohttp.ResponseWriter, r *gohttp.Request) {
			w.Write(errResp)
		})

		server := gohttp.Server{
			Addr:         fmt.Sprintf(":%d", config.Conf.Tracker.HTTP.Port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 7 * time.Second,
			IdleTimeout:  0,
		}
		server.SetKeepAlivesEnabled(false)

		go func() {
			if err := server.ListenAndServe(); err != nil {
				config.Logger.Error("ListenAndServe()", zap.Error(err))
			}
		}()
	}

	// UDP tracker
	if config.Conf.Tracker.UDP.Enabled {
		config.Logger.Info("udp tracker enabled", zap.Int("port", config.Conf.Tracker.UDP.Port))
		udptracker.Init(peerdb)
		go udptracker.Serve()
	}

	if config.Conf.Debug.ExpvarInterval > 0 {
		publishExpvar(peerdb, &httptracker, &udptracker)
	} else {
		select {}
	}
}
