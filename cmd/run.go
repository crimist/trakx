package cmd

import (
	"expvar"
	"fmt"
	"net"
	gohttp "net/http"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/tracker"
	"github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/udp"
	"github.com/crimist/trakx/tracker/udp/connections"
	"go.uber.org/zap"

	// import database types so init is called
	"github.com/crimist/trakx/storage/inmemory"
)

// Run initializes and runs the tracker with the requested configuration settings.
func Run(conf *config.Configuration) {
	type serveConfig struct {
		ip       net.IP
		port     int
		routines int
	}

	var trackers []tracker.Tracker
	var serveConfigs []serveConfig
	var connectionsDB *connections.Connections
	var err error

	zap.L().Debug("Starting Trakx")

	if conf.Stats.General {
		if conf.Stats.Interval <= 0 {
			zap.L().Fatal("Invalid configuration: Stats.Interval must be greater than 0 if Stats.General is enabled")
		}

		zap.L().Info("Collecting statistics", zap.Bool("general", conf.Stats.General), zap.Bool("ip", conf.Stats.IP), zap.Duration("interval", conf.Stats.Interval))

		if conf.HTTP.Mode == config.TrackerModeDisabled {
			zap.L().Warn("Statistics collection enabled but no HTTP server is enabled to publish them")
		}
	}

	// TODO: cache the prev IP collector map size :D
	collector := stats.NewCollectors(conf.Stats.General, conf.Stats.IP, 0)

	db, err := inmemory.NewInMemory(inmemory.Config{
		InitalSize:         0, // TODO: cache this on exit and load on startup
		Persistance:        &inmemory.FilePersistance{},
		PersistanceAddress: conf.DB.Backup.Path,
		EvictionFrequency:  conf.DB.Trim,
		ExpirationTime:     conf.DB.Expiry,
		Collector:          collector,
	})

	if err != nil {
		zap.L().Fatal("Failed to initialize database", zap.Error(err))
	} else {
		zap.L().Info("Initialized database", zap.Int("torrents", db.Torrents()))
	}

	if conf.UDP.Enabled {
		zap.L().Info("UDP tracker enabled", zap.String("ip", conf.UDP.IP), zap.Int("port", conf.UDP.Port))

		connectionsDB = connections.NewConnections(0, conf.UDP.ConnDB.Expiry, conf.UDP.ConnDB.Trim)

		trackers = append(trackers, udp.NewTracker(db, connectionsDB, collector, tracker.TrackerConfig{
			Validate:         conf.UDP.ConnDB.Validate,
			DefaultNumwant:   conf.Numwant.Default,
			MaximumNumwant:   conf.Numwant.Limit,
			Interval:         uint(conf.Announce.Base),
			IntervalVariance: uint(conf.Announce.Fuzz),
		}))

		ip := net.ParseIP(conf.UDP.IP)
		if conf.UDP.IP != "" && ip == nil {
			zap.L().Fatal("Invalid IP address for UDP tracker", zap.String("ip", conf.UDP.IP))
		}

		serveConfigs = append(serveConfigs, serveConfig{
			ip:       ip,
			port:     conf.UDP.Port,
			routines: conf.UDP.Threads,
		})
	}

	// TODO: make http tracker enable based on port with -1 for disabled
	// info mode will run if stats are enabled and/or fileserver mode and tracker mode is disabled
	// ah but then how to set port -.- xd
	if conf.HTTP.Mode == config.TrackerModeEnabled {
		zap.L().Info("HTTP tracker enabled", zap.String("ip", conf.HTTP.IP), zap.Int("port", conf.HTTP.Port))

		// TODO: HTTP serve path in config, also validate it here (ie. dir exists and can access)
		// docs = leave serve path blank to disable serving files
		trackers = append(trackers, http.NewTracker(db, "/TODO/", collector, tracker.TrackerConfig{
			DefaultNumwant:   conf.Numwant.Default,
			MaximumNumwant:   conf.Numwant.Limit,
			Interval:         uint(conf.Announce.Base),
			IntervalVariance: uint(conf.Announce.Fuzz),
			ReadTimeout:      conf.HTTP.Timeout.Read,
			WriteTimeout:     conf.HTTP.Timeout.Write,
		}))

		ip := net.ParseIP(conf.HTTP.IP)
		if conf.HTTP.IP != "" && ip == nil {
			zap.L().Fatal("Invalid IP address for HTTP tracker", zap.String("ip", conf.HTTP.IP))
		}

		serveConfigs = append(serveConfigs, serveConfig{
			ip:       ip,
			port:     conf.HTTP.Port,
			routines: conf.HTTP.Threads,
		})
	} else if conf.HTTP.Mode == config.TrackerModeInfo {
		mux := gohttp.NewServeMux()
		if conf.Stats.General {
			mux.HandleFunc("/stats", func(w gohttp.ResponseWriter, r *gohttp.Request) {
				expvar.Handler().ServeHTTP(w, r)
			})
		}
		mux.Handle("/", gohttp.FileServer(gohttp.Dir("TODO")))

		server := gohttp.Server{
			Addr:         fmt.Sprintf(":%d", conf.HTTP.Port),
			Handler:      mux,
			ReadTimeout:  conf.HTTP.Timeout.Read,
			WriteTimeout: conf.HTTP.Timeout.Write,
			IdleTimeout:  0,
		}
		server.SetKeepAlivesEnabled(false)

		zap.L().Info("Running HTTP web server", zap.Int("port", conf.HTTP.Port))
		go func() {
			if err := server.ListenAndServe(); err != nil {
				zap.L().Fatal("Failed to start HTTP web server", zap.Error(err))
			}
		}()
	}

	go signalHandler(db, trackers)

	if conf.Stats.General {
		go stats.PublishPeriodic(stats.PeriodicConfig{
			Collector: collector,
			Interval:  conf.Stats.Interval,
			GetHashes: db.Torrents,
			GetConns: func() int {
				if connectionsDB != nil {
					return connectionsDB.Entries()
				}
				return 0
			}})
	}

	if conf.Debug.Pprof != 0 {
		go servePprof(conf.Debug.Pprof)
	}

	for i, t := range trackers {
		go func(i int, t tracker.Tracker) {
			if err := t.Serve(serveConfigs[i].ip, serveConfigs[i].port, serveConfigs[i].routines); err != nil {
				zap.L().Fatal("Failed to serve tracker", zap.Error(err), zap.String("ip", serveConfigs[i].ip.String()), zap.Int("port", serveConfigs[i].port))
			}
		}(i, t)
	}

	select {}
}
