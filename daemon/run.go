package daemon

import (
	"expvar"
	"fmt"
	"io"
	"net"
	gohttp "net/http"
	"os"
	"time"

	"github.com/crimist/trakx/config"
	"github.com/crimist/trakx/pools"
	"github.com/crimist/trakx/stats"
	"github.com/crimist/trakx/tracker"
	"github.com/crimist/trakx/tracker/http"
	"github.com/crimist/trakx/tracker/udp"
	"github.com/crimist/trakx/tracker/udp/connections"
	"github.com/crimist/trakx/utils"
	"go.uber.org/zap"

	// import database types so init is called
	"github.com/crimist/trakx/storage/database"
)

type RunOptions struct {
	ImportReader io.Reader
}

// Run initializes and runs the tracker with the requested configuration settings.
func Run(conf *config.Configuration) {
	RunWithOptions(conf, RunOptions{})
}

// RunWithOptions initializes and runs the tracker with the requested configuration settings.
func RunWithOptions(conf *config.Configuration, opts RunOptions) {
	type serveConfig struct {
		ip       net.IP
		port     int
		routines int
	}

	var trackers []tracker.Tracker
	var serveConfigs []serveConfig
	var connectionsDB *connections.Connections
	var connSnapshot []byte
	var err error

	zap.L().Debug("Starting Trakx")

	pools.Initialize(int(conf.Numwant.Limit))

	baseIntervalSeconds := uint(conf.Announce.Base / time.Second)
	fuzzIntervalSeconds := uint(conf.Announce.Fuzz / time.Second)

	// TODO: cache the prev IP collector map size :D
	collector := stats.NewCollectors(conf.Stats.General, conf.Stats.IP, 0)

	var importReader io.Reader
	if opts.ImportReader != nil {
		connSnapshot, importReader, err = splitCombinedSnapshot(opts.ImportReader)
		if err != nil {
			if closer, ok := opts.ImportReader.(io.Closer); ok {
				closer.Close()
			}
			zap.L().Warn("Failed to load snapshot from import stream", zap.Error(err))
			importReader = nil
			connSnapshot = nil
		}
	} else if conf.DB.Backup.Path != "" {
		backupFile, err := os.Open(conf.DB.Backup.Path)
		if err != nil {
			if os.IsNotExist(err) {
				zap.L().Debug("Database backup file does not exist", zap.String("path", conf.DB.Backup.Path))
			} else {
				zap.L().Warn("Failed to open database backup file", zap.String("path", conf.DB.Backup.Path), zap.Error(err))
			}
		} else {
			connSnapshot, importReader, err = splitCombinedSnapshot(backupFile)
			if err != nil {
				backupFile.Close()
				zap.L().Warn("Failed to load combined snapshot", zap.String("path", conf.DB.Backup.Path), zap.Error(err))
				importReader = nil
				connSnapshot = nil
			}
		}
	}

	db, err := database.NewDatabase(database.Config{
		InitalSize:          0, // TODO: cache this on exit and load on startup
		EvictionFrequency:   conf.DB.GC,
		ExpirationTime:      conf.DB.Expiry,
		Collector:           collector,
		ImportReader:        importReader,
	})

	if err != nil {
		zap.L().Fatal("Failed to initialize database", zap.Error(err))
	} else {
		zap.L().Info("Initialized database", zap.Int("torrents", db.Torrents()))
	}

	if conf.UDP.Port != 0 {
		zap.L().Debug("UDP tracker enabled", zap.String("ip", conf.UDP.IP), zap.Int("port", conf.UDP.Port))

		connectionsDB = connections.NewConnections(0, conf.UDP.Connections.Expiry, conf.UDP.Connections.GC)
		if len(connSnapshot) > 0 {
			if err := connectionsDB.Unmarshal(connSnapshot); err != nil {
				zap.L().Warn("Failed to load UDP connections from snapshot", zap.Error(err))
			} else {
				zap.L().Info("Loaded UDP connections from snapshot", zap.Int("connections", connectionsDB.Entries()))
			}
		}

		trackers = append(trackers, udp.NewTracker(db, tracker.TrackerConfig{
			DefaultNumwant:   conf.Numwant.Default,
			MaximumNumwant:   conf.Numwant.Limit,
			Interval:         baseIntervalSeconds,
			IntervalVariance: fuzzIntervalSeconds,
		}, collector, connectionsDB, conf.UDP.Connections.Validate))

		ip := net.ParseIP(conf.UDP.IP)
		if conf.UDP.IP != "" && ip == nil {
			zap.L().Fatal("Invalid IP address for UDP tracker", zap.String("ip", conf.UDP.IP))
		}

		serveConfigs = append(serveConfigs, serveConfig{
			ip:       ip,
			port:     conf.UDP.Port,
			routines: conf.UDP.Routines,
		})
	}

	if conf.HTTP.Tracker {
		zap.L().Debug("HTTP tracker enabled", zap.String("ip", conf.HTTP.IP), zap.Int("port", conf.HTTP.Port))

		trackers = append(trackers, http.NewTracker(db, tracker.TrackerConfig{
			DefaultNumwant:   conf.Numwant.Default,
			MaximumNumwant:   conf.Numwant.Limit,
			Interval:         baseIntervalSeconds,
			IntervalVariance: fuzzIntervalSeconds,
		}, collector, conf.HTTP.Serve, conf.HTTP.Timeout.Read, conf.HTTP.Timeout.Write))

		ip := net.ParseIP(conf.HTTP.IP)
		if conf.HTTP.IP != "" && ip == nil {
			zap.L().Fatal("Invalid IP address for HTTP tracker", zap.String("ip", conf.HTTP.IP))
		}

		serveConfigs = append(serveConfigs, serveConfig{
			ip:       ip,
			port:     conf.HTTP.Port,
			routines: conf.HTTP.Routines,
		})
	} else if conf.HTTP.Port != 0 {
		mux := gohttp.NewServeMux()
		if conf.Stats.General {
			mux.HandleFunc("/stats", func(w gohttp.ResponseWriter, r *gohttp.Request) {
				expvar.Handler().ServeHTTP(w, r)
			})
		}
		mux.Handle("/", gohttp.FileServer(gohttp.Dir(conf.HTTP.Serve)))

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

	go signalHandler(trackers, func() error {
		return persistSnapshot(db, connectionsDB, conf.DB.Backup.Path)
	})

	if conf.DB.Backup.Interval > 0 && conf.DB.Backup.Path != "" {
		zap.L().Debug("Combined backup on interval", zap.Duration("interval", conf.DB.Backup.Interval))
		go utils.RunOn(conf.DB.Backup.Interval, func() {
			if err := writeCombinedSnapshotFile(db, connectionsDB, conf.DB.Backup.Path); err != nil {
				zap.L().Error("failed to write combined backup on interval", zap.Error(err))
			}
		})
	}

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
