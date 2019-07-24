package tracker

import (
	"fmt"
	"net/http"
	"time"

	httptracker "github.com/Syc0x00/Trakx/tracker/http"
	"github.com/Syc0x00/Trakx/tracker/shared"
	udptracker "github.com/Syc0x00/Trakx/tracker/udp"
	"go.uber.org/zap"
)

// Run runs the tracker
func Run() {
	// Init shared stuff
	if err := shared.Init(); err != nil {
		panic(err)
	}

	go handleSigs()
	if shared.Config.ExpvarPort != 0 {
		go Expvar()
	}

	// HTTP tracker / routes
	initRoutes()

	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)
	trackerMux.HandleFunc("/stats", stats)

	if shared.Config.HTTPTracker {
		shared.Logger.Info("http tracker on")
		trackerMux.HandleFunc("/scrape", httptracker.ScrapeHandle)
		trackerMux.HandleFunc("/announce", httptracker.AnnounceHandle)
	} else {
		emptyHandler := func(w http.ResponseWriter, r *http.Request) {}
		trackerMux.HandleFunc("/scrape", emptyHandler)
		trackerMux.HandleFunc("/announce", emptyHandler)
	}

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", shared.Config.HTTPPort),
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
	if shared.Config.UDPPort != 0 {
		shared.Logger.Info("udp tracker on")
		udptracker.Run(time.Duration(shared.Config.UDPTrim) * time.Second)
	}

	select {} // block forever
}
