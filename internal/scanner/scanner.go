package scanner

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	Port    int
	Timeout time.Duration
	Mode    Mode
	SNI     string
}

type Hit struct {
	IP      string
	Latency time.Duration
}

func ScanWorker(ctx context.Context, wg *sync.WaitGroup, cfg Config, jobs <-chan string, progress *atomic.Uint64, hits chan<- Hit) {
	defer wg.Done()
	for {
		select {
		case ip, ok := <-jobs:
			if !ok {
				return
			}
			if latency, ok := Probe(ctx, ip, cfg.Port, cfg.Timeout, cfg.Mode, cfg.SNI); ok {
				select {
				case hits <- Hit{IP: ip, Latency: latency}:
				case <-ctx.Done():
					return
				}
			}
			progress.Add(1)
		case <-ctx.Done():
			return
		}
	}
}
