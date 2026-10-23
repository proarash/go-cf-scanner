package report

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/proarash/go-cf-scanner/internal/scanner"
)

// SaveHits appends every hit to hits.txt as soon as it arrives, calls onHit, and returns all hits
// sorted by latency on done once the hits channel is closed.
func SaveHits(hits <-chan scanner.Hit, onHit func(), done chan<- []scanner.Hit) {
	var all []scanner.Hit
	defer func() {
		sort.Slice(all, func(i, j int) bool { return all[i].Latency < all[j].Latency })
		done <- all
	}()

	file, err := os.OpenFile("hits.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	if file != nil {
		defer file.Close()
	}

	for hit := range hits {
		all = append(all, hit)
		if file != nil {
			if _, err := fmt.Fprintf(file, "%s %d\n", hit.IP, hit.Latency.Milliseconds()); err != nil {
				log.Println(err)
			}
		}
		onHit()
	}
}
