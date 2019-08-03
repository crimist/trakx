package http

import (
	"github.com/syc0x00/trakx/tracker/shared"
	"go.uber.org/zap"
)

type HTTPTracker struct {
	conf   *shared.Config
	logger *zap.Logger
	peerdb *shared.PeerDatabase
}

func NewHTTPTracker(conf *shared.Config, logger *zap.Logger, peerdb *shared.PeerDatabase) *HTTPTracker {
	tracker := HTTPTracker{
		conf:   conf,
		logger: logger,
		peerdb: peerdb,
	}

	return &tracker
}
