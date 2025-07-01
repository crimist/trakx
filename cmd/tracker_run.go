package tracker

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
	var err error

	zap.L().Debug("Starting Trakx")

	warnings := conf.Validate()
	if warnings&config.WarningUDPValidation != 0 {
		zap.L().Warn("Configuration warning [UDP.ConnDB.Validate]: UDP connection validation is disabled. Do not expose this service to untrusted networks; it could be abused in UDP based amplification attacks.")
	}
	if warnings&config.WarningPeerExpiry != 0 {
		zap.L().Warn("Configuration warning [conf.Announce]: Peer expiry time < announce interval. Peers will expire from the database between announces")
	}

	// TODO: conf for collector enable/disable + cache the prev IP collector map size :D
	collector := stats.NewCollectors(false, false, 0)

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

		connectionsDB := connections.NewConnections(0, conf.UDP.ConnDB.Expiry, conf.UDP.ConnDB.Trim)

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

	// TODO: HTTP tracker
	// TODO: migrate these to use port == -1 to disable
	// and then to enable the info mode maybe a seperate boolean option
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
		mux.HandleFunc("/stats", func(w gohttp.ResponseWriter, r *gohttp.Request) {
			expvar.Handler().ServeHTTP(w, r)
		})
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

	if conf.ExpvarInterval > 0 {
		// TODO: publish expvars
		// stats.Publish(peerdb, func() int64 {
		// 	return int64(udptracker.Connections())
		// })

		select {}
	} else {
		zap.L().Debug("Finished Run() no expvar - blocking forever")
		select {}
	}
}
