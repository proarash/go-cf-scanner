package report

import (
	"fmt"
	"sync/atomic"
	"time"
)

func UpdateProgress(total uint64, progress, hitCount *atomic.Uint64, done <-chan struct{}) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	print := func() {
		current := progress.Load()
		percent := float64(current) * 100 / float64(max(total, 1))
		fmt.Printf("\r[Progress]: %.2f%% | [Count]: %d | [Hits]: %d | Press CTRL + C to stop", percent, current, hitCount.Load())
	}

	for {
		select {
		case <-done:
			print()
			fmt.Println()
			return
		case <-ticker.C:
			print()
		}
	}
}
