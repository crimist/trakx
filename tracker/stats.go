package tracker

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

// !!! Not planning to implement. easier to link a netdata bage

// Stats provides information for the index page
type Stats struct {
	Directory string
}

// Active Gets number of active peers/seeders
func (s *Stats) Active() ([]byte, error) {
	var xval []time.Time
	var seeds []float64
	var peers []float64

	for i := 0; i < 100; i++ {
		xval = append(xval, time.Now().Add(time.Duration(-i)*time.Minute))

		val := 0
		for val < 3 {
			val = rand.Intn(6)
		}
		seeds = append(seeds, float64(100*val))
		peers = append(peers, float64(100*val+100))
	}

	graph := chart.Chart{
		Width:  2560,
		Height: 1440,

		XAxis: chart.XAxis{
			Style:          chart.StyleShow(),
			ValueFormatter: chart.TimeMinuteValueFormatter,
			Name:           "Time",
			NameStyle: chart.Style{
				Show: true,
			},
		},
		YAxis: chart.YAxis{
			Style: chart.StyleShow(),
			Name:  "Peers/Seeds",
			NameStyle: chart.Style{
				Show: true,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Style: chart.Style{
					Show:        true,
					FillColor:   drawing.ColorBlue,
					StrokeColor: drawing.ColorBlue,
				},
				XValues: xval,
				YValues: peers,
			},
			chart.TimeSeries{
				Style: chart.Style{
					Show:        true,
					StrokeColor: drawing.ColorGreen,
				},
				XValues: xval,
				YValues: seeds,
			},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		panic(err)
	}

	return buffer.Bytes(), nil
}

// Generator will continusly generate stats graphs to the given web dir
func (s *Stats) Generator() {
	for c := time.Tick(1 * time.Minute); ; <-c {
		activeGraph, err := s.Active()
		if err != nil {
			logger.Error(err.Error())
		}

		err = ioutil.WriteFile(s.Directory+"active.png", activeGraph, 644)
		if err != nil {
			logger.Error(err.Error())
		}
	}
}
