package tracker

import (
	"net/http"
	"time"

	httptracker "github.com/Syc0x00/Trakx/tracker/http"
	"github.com/Syc0x00/Trakx/tracker/shared"
	udptracker "github.com/Syc0x00/Trakx/tracker/udp"
	"github.com/thoas/stats"
)

// Run runs the tracker
func Run(prod bool) {
	// Init shared stuff
	if err := shared.Init(prod); err != nil {
		panic(err)
	}

	// HTTP tracker / routes
	statsMiddleware := stats.New()
	trackerMux := http.NewServeMux()
	trackerMux.HandleFunc("/", index)
	trackerMux.HandleFunc("/dmca", dmca)
	trackerMux.HandleFunc("/scrape", httptracker.ScrapeHandle)
	trackerMux.HandleFunc("/announce", httptracker.AnnounceHandle)

	server := http.Server{
		Addr:         ":" + shared.HTTPPort,
		Handler:      statsMiddleware.Handler(trackerMux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	// UDP tracker
	udp := udptracker.UDPTracker{}
	go udp.Trimmer()

	udp.Listen()
}
