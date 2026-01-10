package main

import (
	"context"
	"fmt"
	mrand "math/rand"
	"net"
)

func seedPhase(ctx context.Context, cfg config, ds *dataset) error {
	if cfg.seedTorrents == 0 {
		return nil
	}

	rng := mrand.New(mrand.NewSource(cfg.rngSeed + 11))
	if cfg.seedPeersDist == "fixed" && cfg.seedPeersPerTor <= 0 {
		return fmt.Errorf("seed-peers-per-torrent must be > 0 for fixed distribution")
	}

	seedHashes := ds.torrents
	if cfg.seedTorrents > 0 && cfg.seedTorrents < len(seedHashes) {
		seedHashes = seedHashes[:cfg.seedTorrents]
	}

	peersPerTorrent := make([]int, len(seedHashes))
	seedPeersTotal := 0
	for i := range peersPerTorrent {
		count, err := cfg.seedPeersForTorrent(rng)
		if err != nil {
			return err
		}
		peersPerTorrent[i] = count
		seedPeersTotal += count
	}

	seedData := newDataset(rng, 1, seedPeersTotal)
	seedPeerIDs := seedData.peers
	seedEncodedPeers := seedData.encodedPeers
	seedNumwant := cfg.numwant
	if seedNumwant < 0 {
		seedNumwant = 0
	}
	switch cfg.mode {
	case "udp":
		addr, err := net.ResolveUDPAddr("udp", cfg.udpAddr)
		if err != nil {
			return err
		}
		client, err := newUDPWorker(addr, cfg.timeout, cfg.udpConnRefresh)
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
				if err := client.announce(rng, hash, peer, 1000, int32(seedNumwant), uint16(defaultPort)); err != nil {
					return err
				}
				peerIdx++
			}
		}
	case "http":
		client := newHTTPWorker(cfg.httpAddr, cfg.timeout, cfg.httpHostHeader)
		peerIdx := 0
		for t := 0; t < len(seedHashes); t++ {
			for p := 0; p < peersPerTorrent[t]; p++ {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				peer := seedEncodedPeers[peerIdx%len(seedEncodedPeers)]
				hash := ds.encodedHashes[t%len(ds.encodedHashes)]
				payload := buildHTTPAnnounce(hash, peer, cfg.httpHostHeader, defaultPort, 1000, seedNumwant, cfg.compact)
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
