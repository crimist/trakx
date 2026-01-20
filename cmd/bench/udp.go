package main

import (
	"context"
	"encoding/binary"
	"fmt"
	mrand "math/rand"
	"net"
	"time"

	"github.com/crimist/trakx/internal/tracker/udp/udpprotocol"
)

type udpWorker struct {
	addr         *net.UDPAddr
	conn         *net.UDPConn
	connID       int64
	nextRefresh  time.Time
	refreshEvery time.Duration
	timeout      time.Duration
}

func newUDPWorker(addr *net.UDPAddr, timeout, refresh time.Duration) (*udpWorker, error) {
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	return &udpWorker{
		addr:         addr,
		conn:         conn,
		timeout:      timeout,
		refreshEvery: refresh,
	}, nil
}

func (w *udpWorker) close() {
	if w.conn != nil {
		w.conn.Close()
	}
}

func (w *udpWorker) ensureConnID(rng *mrand.Rand) error {
	now := time.Now()
	if w.connID != 0 && (w.refreshEvery <= 0 || now.Before(w.nextRefresh)) {
		return nil
	}
	tx := rng.Int31()
	req := make([]byte, 16)
	binary.BigEndian.PutUint64(req[0:8], uint64(udpprotocol.ProtocolMagic))
	binary.BigEndian.PutUint32(req[8:12], uint32(udpprotocol.ActionConnect))
	binary.BigEndian.PutUint32(req[12:16], uint32(tx))
	w.conn.SetWriteDeadline(time.Now().Add(w.timeout))
	if _, err := w.conn.Write(req); err != nil {
		return err
	}
	resp := make([]byte, 16)
	w.conn.SetReadDeadline(time.Now().Add(w.timeout))
	n, err := w.conn.Read(resp)
	if err != nil {
		return err
	}
	if n < 16 {
		return fmt.Errorf("short connect response: %d", n)
	}
	action := binary.BigEndian.Uint32(resp[0:4])
	if action == uint32(udpprotocol.ActionError) {
		return fmt.Errorf("connect error response")
	}
	if action != uint32(udpprotocol.ActionConnect) {
		return fmt.Errorf("unexpected connect action: %d", action)
	}
	respTx := int32(binary.BigEndian.Uint32(resp[4:8]))
	if respTx != tx {
		return fmt.Errorf("connect transaction mismatch")
	}
	w.connID = int64(binary.BigEndian.Uint64(resp[8:16]))
	if w.refreshEvery > 0 {
		w.nextRefresh = now.Add(w.refreshEvery)
	}
	return nil
}

func (w *udpWorker) announce(rng *mrand.Rand, infoHash, peerID []byte, left int64, numwant int32, port uint16) error {
	tx := rng.Int31()
	buf := make([]byte, 98)
	binary.BigEndian.PutUint64(buf[0:8], uint64(w.connID))
	binary.BigEndian.PutUint32(buf[8:12], uint32(udpprotocol.ActionAnnounce))
	binary.BigEndian.PutUint32(buf[12:16], uint32(tx))
	copy(buf[16:36], infoHash)
	copy(buf[36:56], peerID)
	binary.BigEndian.PutUint64(buf[56:64], 0)
	binary.BigEndian.PutUint64(buf[64:72], uint64(left))
	binary.BigEndian.PutUint64(buf[72:80], 0)
	binary.BigEndian.PutUint32(buf[80:84], uint32(udpprotocol.EventNone))
	binary.BigEndian.PutUint32(buf[84:88], 0)
	binary.BigEndian.PutUint32(buf[88:92], 0)
	binary.BigEndian.PutUint32(buf[92:96], uint32(numwant))
	binary.BigEndian.PutUint16(buf[96:98], port)
	w.conn.SetWriteDeadline(time.Now().Add(w.timeout))
	if _, err := w.conn.Write(buf); err != nil {
		return err
	}
	resp := make([]byte, 2048)
	w.conn.SetReadDeadline(time.Now().Add(w.timeout))
	n, err := w.conn.Read(resp)
	if err != nil {
		return err
	}
	if n < 8 {
		return fmt.Errorf("short announce response: %d", n)
	}
	action := binary.BigEndian.Uint32(resp[0:4])
	if action == uint32(udpprotocol.ActionError) {
		return fmt.Errorf("announce error response")
	}
	respTx := int32(binary.BigEndian.Uint32(resp[4:8]))
	if respTx != tx {
		return fmt.Errorf("announce transaction mismatch")
	}
	return nil
}

func (w *udpWorker) scrape(rng *mrand.Rand, hashes [][]byte) error {
	tx := rng.Int31()
	buf := make([]byte, 16+len(hashes)*20)
	binary.BigEndian.PutUint64(buf[0:8], uint64(w.connID))
	binary.BigEndian.PutUint32(buf[8:12], uint32(udpprotocol.ActionScrape))
	binary.BigEndian.PutUint32(buf[12:16], uint32(tx))
	pos := 16
	for _, h := range hashes {
		copy(buf[pos:pos+20], h)
		pos += 20
	}
	w.conn.SetWriteDeadline(time.Now().Add(w.timeout))
	if _, err := w.conn.Write(buf); err != nil {
		return err
	}
	resp := make([]byte, 2048)
	w.conn.SetReadDeadline(time.Now().Add(w.timeout))
	n, err := w.conn.Read(resp)
	if err != nil {
		return err
	}
	if n < 8 {
		return fmt.Errorf("short scrape response: %d", n)
	}
	action := binary.BigEndian.Uint32(resp[0:4])
	if action == uint32(udpprotocol.ActionError) {
		return fmt.Errorf("scrape error response")
	}
	respTx := int32(binary.BigEndian.Uint32(resp[4:8]))
	if respTx != tx {
		return fmt.Errorf("scrape transaction mismatch")
	}
	return nil
}

type udpBenchWorker struct {
	id     int
	client *udpWorker
}

func (w *udpBenchWorker) run(ctx context.Context, cfg config, ds *dataset, limiter *rateLimiter) *workerMetrics {
	metrics := newWorkerMetrics()
	rng := mrand.New(mrand.NewSource(cfg.rngSeed + int64(w.id*rngOffsetUDP)))
	peerIdx := w.id % len(ds.peers)
	peerID := ds.peers[peerIdx]

	left := int64(1000)
	if rng.Float64() < defaultSeedFraction {
		left = 0
	}

	var hashIndexes []int
	for i := 0; i < defaultTorrentsPerPeer; i++ {
		hashIndexes = append(hashIndexes, rng.Intn(len(ds.torrents)))
	}

	for {
		select {
		case <-ctx.Done():
			return metrics
		default:
		}
		if limiter != nil {
			select {
			case <-ctx.Done():
				return metrics
			case <-limiter.tokens:
			}
		}
		if err := w.client.ensureConnID(rng); err != nil {
			metrics.errors++
			continue
		}

		reqType := requestAnnounce
		if rng.Float64() < defaultScrapeRatio {
			reqType = requestScrape
		}

		start := time.Now()
		var err error
		if reqType == requestAnnounce {
			numwant := pickNumwant(rng)
			hash := ds.torrents[hashIndexes[rng.Intn(len(hashIndexes))]]
			err = w.client.announce(rng, hash, peerID, left, int32(numwant), uint16(defaultPort))
		} else {
			hashes := make([][]byte, defaultScrapeHashes)
			for i := 0; i < defaultScrapeHashes; i++ {
				hashes[i] = ds.torrents[rng.Intn(len(ds.torrents))]
			}
			err = w.client.scrape(rng, hashes)
		}
		lat := time.Since(start)
		metrics.counts[reqType]++
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				metrics.timeouts++
			} else {
				metrics.errors++
			}
			continue
		}
		metrics.success++
		metrics.hist.observe(lat)
	}
}
