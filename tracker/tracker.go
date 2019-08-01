package tracker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/syc0x00/trakx/bencoding"
	httptracker "github.com/syc0x00/trakx/tracker/http"
	"github.com/syc0x00/trakx/tracker/shared"
	udptracker "github.com/syc0x00/trakx/tracker/udp"
	"go.uber.org/zap"
)

// Run runs the tracker
func Run() {
	if err := shared.Init(); err != nil {
		panic(err)
	}

	go handleSigs()
	if shared.Config.Trakx.Expvar.Enabled {
		go publishExpvar()
	}

	// HTTP tracker / routes
	initRoutes()

	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)
	trackerMux.HandleFunc("/stats", stats)

	if shared.Config.Tracker.HTTP.Enabled {
		shared.Logger.Info("http tracker enabled")
		trackerMux.HandleFunc("/scrape", httptracker.ScrapeHandle)
		trackerMux.HandleFunc("/announce", httptracker.AnnounceHandle)
	} else {
		d := bencoding.NewDict()
		d.Add("interval", 432000) // 5 days
		errResp := []byte(d.Get())

		trackerMux.HandleFunc("/scrape", func(w http.ResponseWriter, r *http.Request) {})
		trackerMux.HandleFunc("/announce", func(w http.ResponseWriter, r *http.Request) {
			w.Write(errResp)
		})
	}

	server := http.Server{
		Addr:         fmt.Sprintf(":%d", shared.Config.Tracker.HTTP.Port),
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
	if shared.Config.Tracker.UDP.Enabled {
		shared.Logger.Info("udp tracker enabled")
		udptracker.Run()
	}

	select {} // block forever
}
