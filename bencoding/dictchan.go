package bencoding

import "github.com/crimist/trakx/tracker/storage"

// global dictChannel
var dictChan dictChannel

func init() {
	dictChan.Init()
}

type dictChannel struct {
	channel chan *Dictionary
}

func (dc *dictChannel) Init() {
	const max = 1e4

	dictChan.channel = make(chan *Dictionary, max)
}

func (dc *dictChannel) Get() *Dictionary {
	// if empty create new
	if len(dc.channel) == 0 {
		storage.Expvar.Pools.Dict.Add(1)
		return new(Dictionary)
	}

	// otherwise pull off queue
	return <-dc.channel
}

func (dc *dictChannel) Put(d *Dictionary) {
	// if queue is full than discard
	if len(dc.channel) == cap(dc.channel) {
		return
	}

	// otherwise add back to queue
	dc.channel <- d
}
