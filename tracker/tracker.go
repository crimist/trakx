package tracker

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/syc0x00/trakx/bencoding"
	httptracker "github.com/syc0x00/trakx/tracker/http"
	"github.com/syc0x00/trakx/tracker/shared"
	"github.com/syc0x00/trakx/tracker/udp"
	"go.uber.org/zap"
)

var (
	udptracker *udp.UDPTracker
	logger     *zap.Logger
	conf       *shared.Config
	root       string
)

// Run runs the tracker
func Run() {
	var err error

	root, err = os.UserHomeDir()
	if err != nil {
		panic("os.UserHomeDir() failed: " + err.Error())
	}
	root += "/.trakx/"
	conf = shared.NewConfig(root)

	cfg := zap.NewDevelopmentConfig()
	logger, err = cfg.Build()
	if err != nil {
		panic(err)
	}

	peerdb := shared.NewPeerDatabase(conf, logger)
	shared.InitExpvar(peerdb)

	go handleSigs(peerdb)

	// HTTP tracker / routes
	initRoutes()

	if conf.Tracker.HTTP.Enabled {
		logger.Info("http tracker enabled")

		t := httptracker.NewHTTPTracker(conf, logger, peerdb)
		go t.Serve(indexData)
	} else {
		d := bencoding.NewDict()
		d.Add("interval", 432000) // 5 days
		errResp := []byte(d.Get())

		trackerMux := http.NewServeMux()
		trackerMux.HandleFunc("/", index)
		trackerMux.HandleFunc("/dmca", dmca)
		trackerMux.HandleFunc("/stats", stats)
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
		logger.Info("udp tracker enabled")
		udptracker = udp.NewUDPTracker(conf, logger, peerdb)
	}

	if conf.Trakx.Expvar.Enabled {
		go publishExpvar(conf, peerdb)
	}

	select {} // block forever
}
