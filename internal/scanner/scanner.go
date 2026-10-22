package scanner

import (
	"context"
	"sync"
)

func ScanWorker(jobs <-chan string, wg *sync.WaitGroup, timeout int, progress chan<- uint64, hits chan<- string, ctx context.Context, hitCounter chan<- uint8) {
	defer wg.Done()
	for {
		select {
		case ip, ok := <-jobs:
			if !ok {
				return
			}
			valid := GetTcping(ip, 443, timeout, ctx)
			if valid {
				hits <- ip
				select {
				case hitCounter <- 1:
				case <-ctx.Done():
					return
				}
			}
			progress <- 1
		case <-ctx.Done():
			return
		}
	}
}
