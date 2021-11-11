package bencoding

import "github.com/crimist/trakx/tracker/storage"

const dictionaryChanMax = 1e4

var dictionaryChan chan *Dictionary

func init() {
	dictionaryChan = make(chan *Dictionary, dictionaryChanMax)
}

// GetDictionary returns a Dictionary pointer from the Dictionary pool.
func GetDictionary() *Dictionary {
	// if empty create new
	if len(dictionaryChan) == 0 {
		storage.Expvar.Pools.Dict.Add(1)
		return NewDictionary()
	}

	// otherwise pull off queue & init
	d := <-dictionaryChan
	d.Reset()
	return d
}

// GetDictionary puts a Dictionary back in the Dictionary pool.
func PutDictionary(d *Dictionary) {
	// if queue is full than discard
	if len(dictionaryChan) == cap(dictionaryChan) {
		return
	}

	// otherwise add it back
	dictionaryChan <- d
}
