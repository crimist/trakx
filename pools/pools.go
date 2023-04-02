package pools

import (
	"github.com/crimist/trakx/bencoding"
)

var (
	Peerlists4   *Pool[[]byte]
	Peerlists6   *Pool[[]byte]
	Dictionaries *Pool[*bencoding.Dictionary]
)

// TODO: create these pools in their respective packages, most of these don't need to be global variables
func Initialize(numwantLimit int) {
	peerlist4Max := 6 * numwantLimit  // ipv4 + port
	peerlist6Max := 18 * numwantLimit // ipv6 + port

	Peerlists4 = NewPool(func() any {
		data := make([]byte, peerlist4Max)
		return data
	}, func(data []byte) {
		data = data[:peerlist4Max]
		_ = data
	})

	Peerlists6 = NewPool(func() any {
		data := make([]byte, peerlist6Max)
		return data
	}, func(data []byte) {
		data = data[:peerlist6Max]
		_ = data
	})

	Dictionaries = NewPool(func() any {
		return bencoding.NewDictionary()
	}, func(dictionary *bencoding.Dictionary) {
		dictionary.Reset()
	})
}
