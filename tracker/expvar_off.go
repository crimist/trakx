// +build !expvar

package tracker

import (
	"sync/atomic"

	"github.com/syc0x00/trakx/tracker/shared"
)

func publishExpvar(conf interface{}, peerdb interface{}, ht interface{}) {
	atomic.StoreInt64(&shared.Expvar.Announces, 0)
	atomic.StoreInt64(&shared.Expvar.AnnouncesOK, 0)
	atomic.StoreInt64(&shared.Expvar.Scrapes, 0)
	atomic.StoreInt64(&shared.Expvar.ScrapesOK, 0)
	atomic.StoreInt64(&shared.Expvar.Clienterrs, 0)
	atomic.StoreInt64(&shared.Expvar.Connects, 0)
	atomic.StoreInt64(&shared.Expvar.ConnectsOK, 0)
}
