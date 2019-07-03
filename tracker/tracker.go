package tracker

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	httptracker "github.com/Syc0x00/Trakx/tracker/http"
	"github.com/Syc0x00/Trakx/tracker/shared"
	udptracker "github.com/Syc0x00/Trakx/tracker/udp"
	"go.uber.org/zap"
)

// Run runs the tracker
func Run(prod, udpTracker, httpTracker bool) {
	// Init shared stuff
	if err := shared.Init(prod); err != nil {
		panic(err)
	}

	go Expvar()

	// HTTP tracker / routes
	initRoutes()

	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)

	if httpTracker {
		shared.Logger.Info("http tracker on")
		trackerMux.HandleFunc("/scrape", httptracker.ScrapeHandle)
		trackerMux.HandleFunc("/announce", httptracker.AnnounceHandle)
	} else {
		emptyHandler := func(w http.ResponseWriter, r *http.Request) {}
		trackerMux.HandleFunc("/scrape", emptyHandler)
		trackerMux.HandleFunc("/announce", emptyHandler)
	}

	server := http.Server{
		Addr:         ":" + shared.HTTPPort,
		Handler:      trackerMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 7 * time.Second,
		IdleTimeout:  0,
	}
	server.SetKeepAlivesEnabled(false)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			shared.Logger.Error("ListenAndServe()", zap.Error(err))
		}
	}()

	// UDP tracker
	if udpTracker {
		shared.Logger.Info("udp tracker on")
		udptracker.Run(shared.UDPTrimInterval)
	}

	select {} // block forever
}
