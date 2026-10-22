package report

import (
	"context"
	"fmt"
	"log"
	"time"
)

func UpdateProgress(total *uint64, progress <-chan uint64, done chan struct{}, hitCounter <-chan uint8, cancelFunc context.CancelFunc) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	var currentProgress uint64 = 0
	var hitsCount uint8 = 0

	log.Printf("Started scanning total %v IPs", *total)

	for {
		select {
		case <-done:
			log.Printf("Done")
			return
		case value, ok := <-hitCounter:
			if ok {
				hitsCount += value
				if hitsCount == 10 {
					log.Printf("about  >= %v success hits, closing program.\n", hitsCount)
					cancelFunc()
					return
				}
			}
		case value, ok := <-progress:
			if ok {
				currentProgress += value
			}
		case <-ticker.C:
			if *total > 0 {
				current := float32(currentProgress) * 100.0 / float32(*total)
				fmt.Printf("\r[Progress]: %.2f%% | [Count]: %d | Press CTRL + C to stop the program", current, currentProgress)
			}
		}
	}
}
