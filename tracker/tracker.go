package tracker

import (
	"expvar"
	"fmt"
	"math/rand"
	gohttp "net/http"
	"time"

	"github.com/crimist/trakx/bencoding"
	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/tracker/config"
	"github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/stats"
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

	if !config.Config.Loaded() {
		config.Logger.Fatal("Config failed to load critical values", zap.Any("config", config.Config))
	}

	config.Logger.Info("Loaded configuration, starting trakx...")

	// configuration warnings
	if !config.Config.UDP.ConnDB.Validate {
		config.Logger.Warn("UDP connection validation is DISABLED. Do not expose to public, sever could be abused for UDP amplication DoS.")
	}
	if config.Config.DB.Expiry < config.Config.Announce.Base+config.Config.Announce.Fuzz {
		// likely a configuration error
		config.Logger.Error("Peer expiry < announce interval. Peers will expire before being updated.")
	}

	// db
	peerdb, err := storage.Open()
	if err != nil {
		config.Logger.Fatal("Failed to initialize storage", zap.Error(err))
	} else {
		config.Logger.Info("Initialized storage")
	}

	pools.Initialize(int(config.Config.Numwant.Limit))

	// run signal handler
	go signalHandler(peerdb, &udptracker, &httptracker)

	// run pprof server
	if config.Config.Debug.Pprof != 0 {
		go servePprof()
	}

	if config.Config.HTTP.Mode == config.TrackerModeEnabled {
		config.Logger.Info("HTTP tracker enabled", zap.Int("port", config.Config.HTTP.Port), zap.String("ip", config.Config.HTTP.IP))

		httptracker.Init(peerdb)
		go func() {
			if err := httptracker.Serve(); err != nil {
				config.Logger.Fatal("Failed to serve HTTP tracker", zap.Error(err))
			}
		}()
	} else if config.Config.HTTP.Mode == config.TrackerModeInfo {
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
		mux.HandleFunc("/heartbeat", func(w gohttp.ResponseWriter, r *gohttp.Request) {})
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
			Addr:         fmt.Sprintf(":%d", config.Config.HTTP.Port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 7 * time.Second,
			IdleTimeout:  0,
		}
		server.SetKeepAlivesEnabled(false)

		config.Logger.Info("Running HTTP info server", zap.Int("port", config.Config.HTTP.Port))
		go func() {
			if err := server.ListenAndServe(); err != nil {
				config.Logger.Error("Failed to start HTTP server", zap.Error(err))
			}
		}()
	}

	// UDP tracker
	if config.Config.UDP.Enabled {
		config.Logger.Info("UDP tracker enabled", zap.Int("port", config.Config.UDP.Port), zap.String("ip", config.Config.UDP.IP))
		udptracker.Init(peerdb)

		go func() {
			if err := udptracker.Serve(); err != nil {
				config.Logger.Fatal("Failed to serve UDP tracker", zap.Error(err))
			}
		}()
	}

	if config.Config.ExpvarInterval > 0 {
		stats.Publish(peerdb, func() int64 {
			return int64(udptracker.Connections())
		})
	} else {
		config.Logger.Debug("Finished Run() no expvar - blocking forever")
		select {}
	}
}
