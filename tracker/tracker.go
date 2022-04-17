package tracker

import (
	"expvar"
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

// Run initializes and runs the tracker with the requested configuration settings.
func Run() {
	var udptracker udp.UDPTracker
	var httptracker http.HTTPTracker
	var err error

	rand.Seed(time.Now().UnixNano() * time.Now().Unix())

	if !config.Conf.Loaded() {
		config.Logger.Fatal("Config failed to load critical values", zap.Any("config", config.Conf))
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
		go servePprof()
	}

	if config.Conf.Tracker.HTTP.Mode == config.TrackerModeEnabled {
		config.Logger.Info("HTTP tracker enabled", zap.Int("port", config.Conf.Tracker.HTTP.Port), zap.String("ip", config.Conf.Tracker.HTTP.IP))

		httptracker.Init(peerdb)
		go func() {
			if err := httptracker.Serve(); err != nil {
				config.Logger.Fatal("Failed to serve HTTP tracker", zap.Error(err))
			}
		}()
	} else if config.Conf.Tracker.HTTP.Mode == config.TrackerModeInfo {
		// serve basic html server
		cache, err := config.GenerateEmbeddedCache()
		if err != nil {
			config.Logger.Fatal("failed to generate embedded cache", zap.Error(err))
		}

		// create big interval for announce response to reduce load
		d := bencoding.NewDictionary()
		d.Int64("interval", 86400) // 1 day
		announceResponse := d.GetBytes()

		expvarHandler := expvar.Handler()

		mux := gohttp.NewServeMux()
		mux.HandleFunc("/stats", func(w gohttp.ResponseWriter, r *gohttp.Request) {
			expvarHandler.ServeHTTP(w, r)
		})
		mux.HandleFunc("/scrape", func(w gohttp.ResponseWriter, r *gohttp.Request) {})
		mux.HandleFunc("/announce", func(w gohttp.ResponseWriter, r *gohttp.Request) {
			w.Write(announceResponse)
		})

		for filepath, data := range cache {
			dataBytes := []byte(data)
			mux.HandleFunc(filepath, func(w gohttp.ResponseWriter, r *gohttp.Request) {
				w.Write(dataBytes)
			})
		}

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
				config.Logger.Error("Failed to start HTTP server", zap.Error(err))
			}
		}()
	}

	// UDP tracker
	if config.Conf.Tracker.UDP.Enabled {
		config.Logger.Info("UDP tracker enabled", zap.Int("port", config.Conf.Tracker.UDP.Port), zap.String("ip", config.Conf.Tracker.UDP.IP))
		udptracker.Init(peerdb)

		go func() {
			if err := udptracker.Serve(); err != nil {
				config.Logger.Fatal("Failed to serve UDP tracker", zap.Error(err))
			}
		}()
	}

	if config.Conf.Debug.ExpvarInterval > 0 {
		publishExpvar(peerdb, &httptracker, &udptracker)
	} else {
		config.Logger.Debug("Finished Run() no expvar - blocking forever")
		select {}
	}
}
