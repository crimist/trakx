package tracker

import (
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/Syc0x00/Trakx/bencoding"
	httptracker "github.com/Syc0x00/Trakx/tracker/http"
	"github.com/Syc0x00/Trakx/tracker/shared"
	udptracker "github.com/Syc0x00/Trakx/tracker/udp"
	_ "net/http/pprof"
)

// Run runs the tracker
func Run(prod, udpTracker, httpTracker bool) {
	// Init shared stuff
	if err := shared.Init(prod); err != nil {
		panic(err)
	}

	go Expvar()

	// HTTP tracker / routes
	// TODO: https://groups.google.com/forum/#!topic/golang-nuts/mH3OstyPESA
	initRoutes()

	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)

	if httpTracker {
		shared.Logger.Info("http tracker on")
		trackerMux.HandleFunc("/scrape", httptracker.ScrapeHandle)
		trackerMux.HandleFunc("/announce", httptracker.AnnounceHandle)
	} else {
		dict := bencoding.NewDict()
		dict.Add("failure reason", "Not a tracker")
		dict.Add("retry in", "3600") // 1 hour
		resp := []byte(dict.Get())

		notTracker := func(w http.ResponseWriter, r *http.Request) {
			w.Write(resp)
		}

		trackerMux.HandleFunc("/scrape", notTracker)
		trackerMux.HandleFunc("/announce", notTracker)
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
		udp := udptracker.UDPTracker{}
		go udp.Trimmer()
		go udp.Listen()
	}

	select {}
}
