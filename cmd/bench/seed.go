package main

import (
	"context"
	mrand "math/rand"
	"net"
)

func seedPhase(ctx context.Context, cfg config, ds *dataset) error {
	if cfg.seed == 0 {
		return nil
	}

	rng := mrand.New(mrand.NewSource(cfg.rngSeed + 11))

	seedHashes := ds.torrents
	if cfg.seed > 0 && cfg.seed < len(seedHashes) {
		seedHashes = seedHashes[:cfg.seed]
	}

	// Use lognormal distribution for realistic peer counts.
	peersPerTorrent := make([]int, len(seedHashes))
	seedPeersTotal := 0
	for i := range peersPerTorrent {
		count := seedPeersForTorrent(rng)
		peersPerTorrent[i] = count
		seedPeersTotal += count
	}

	seedData := newDataset(rng, 1, seedPeersTotal)
	seedPeerIDs := seedData.peers
	seedEncodedPeers := seedData.encodedPeers
	seedNumwant := 0 // don't need peers returned

	switch cfg.mode {
	case "udp":
		addr, err := net.ResolveUDPAddr("udp", cfg.target)
		if err != nil {
			return err
		}
		client, err := newUDPWorker(addr, defaultTimeout, defaultConnRefresh)
		if err != nil {
			return err
		}
		defer client.close()
		if err := client.ensureConnID(rng); err != nil {
			return err
		}
		peerIdx := 0
		for t := 0; t < len(seedHashes); t++ {
			for p := 0; p < peersPerTorrent[t]; p++ {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				peer := seedPeerIDs[peerIdx%len(seedPeerIDs)]
				hash := seedHashes[t]
				left := int64(1000)
				if rng.Float64() < defaultSeedFraction {
					left = 0
				}
				if err := client.announce(rng, hash, peer, left, int32(seedNumwant), uint16(defaultPort)); err != nil {
					return err
				}
				peerIdx++
			}
		}
	case "http":
		client := newHTTPWorker(cfg.target, defaultTimeout, cfg.target)
		peerIdx := 0
		for t := 0; t < len(seedHashes); t++ {
			for p := 0; p < peersPerTorrent[t]; p++ {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				peer := seedEncodedPeers[peerIdx%len(seedEncodedPeers)]
				hash := ds.encodedHashes[t%len(ds.encodedHashes)]
				left := int64(1000)
				if rng.Float64() < defaultSeedFraction {
					left = 0
				}
				payload := buildHTTPAnnounce(hash, peer, cfg.target, defaultPort, left, seedNumwant, defaultCompact)
				if err := client.doRequest(payload); err != nil {
					return err
				}
				peerIdx++
			}
		}
	case "both":
		cfgUDP := cfg
		cfgUDP.mode = "udp"
		if err := seedPhase(ctx, cfgUDP, ds); err != nil {
			return err
		}
		cfgHTTP := cfg
		cfgHTTP.mode = "http"
		return seedPhase(ctx, cfgHTTP, ds)
	}
	return nil
}
