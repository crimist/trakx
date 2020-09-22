package bencoding

import "github.com/crimist/trakx/tracker/storage"

var dictChan dictCh

func init() {
	const max = 1e3

	dictChan.channel = make(chan *Dictionary, max)
}

type dictCh struct {
	channel chan *Dictionary
}

func (dc *dictCh) Get() *Dictionary {
	// if empty create new
	if len(dc.channel) == 0 {
		storage.Expvar.Pools.Dict.Add(1)
		return new(Dictionary)
	}

	// otherwise pull off queue
	return <-dc.channel
}

func (dc *dictCh) Put(d *Dictionary) {
	// if queue is full than discard
	if len(dc.channel) == cap(dc.channel) {
		return
	}

	// otherwise add back to queue
	dc.channel <- d
}
