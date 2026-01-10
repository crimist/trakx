package main

import (
	"context"
	mrand "math/rand"
	"net"
)

func seedPhase(ctx context.Context, cfg config, ds *dataset) error {
	if cfg.seedTorrents == 0 || cfg.seedPeersPerTor == 0 {
		return nil
	}
	seedPeers := cfg.seedTorrents * cfg.seedPeersPerTor
	rng := mrand.New(mrand.NewSource(cfg.rngSeed + 11))
	seedData := newDataset(rng, cfg.seedTorrents, seedPeers)
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
		for t := 0; t < cfg.seedTorrents; t++ {
			for p := 0; p < cfg.seedPeersPerTor; p++ {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				peer := seedData.peers[peerIdx%len(seedData.peers)]
				hash := seedData.torrents[t%len(seedData.torrents)]
				if err := client.announce(rng, hash, peer, 1000, int32(cfg.numwant), uint16(defaultPort)); err != nil {
					return err
				}
				peerIdx++
			}
		}
	case "http":
		client := newHTTPWorker(cfg.httpAddr, cfg.timeout, cfg.httpHostHeader)
		peerIdx := 0
		for t := 0; t < cfg.seedTorrents; t++ {
			for p := 0; p < cfg.seedPeersPerTor; p++ {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				peer := seedData.encodedPeers[peerIdx%len(seedData.encodedPeers)]
				hash := seedData.encodedHashes[t%len(seedData.encodedHashes)]
				payload := buildHTTPAnnounce(hash, peer, cfg.httpHostHeader, defaultPort, 1000, cfg.numwant, cfg.compact)
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
