package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"net/netip"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/proarash/go-cf-scanner/internal/cloudflare"
	"github.com/proarash/go-cf-scanner/internal/report"
	"github.com/proarash/go-cf-scanner/internal/scanner"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	bots := 100
	timeout := 3
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	jobs := make(chan string, 2000)
	done := make(chan struct{})
	progress := make(chan uint64, 1000)
	hits := make(chan string, 1000)
	doneHits := make(chan struct{})
	hitCounter := make(chan uint8, 16)

	ips := cloudflare.Ip4Ranges()
	fmt.Println("[0] Randomize and scan all")
	for index, ip := range ips {
		fmt.Printf("[%v] %v\n", index+1, ip)
	}
	fmt.Printf("Select CIDR to scan: ")
	input, err := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if err != nil {
		panic(err)
	}

	log.Printf("Bots: %v | Timeout: %v second(s)\n", bots, timeout)
	selection, err := strconv.ParseInt(input, 0, 8)
	if err != nil {
		panic(err)
	}
	if input == "0" {
		log.Printf("Selected %v, it would shuffle the CIDR then scan all of them.\n", input)
	} else {
		log.Printf("Selected %v | %v\n", input, ips[selection-1])
	}

	var totalIpCount uint64 = 0

	go report.SaveHits(doneHits, hits)

	for range bots {
		wg.Add(1)
		go scanner.ScanWorker(jobs, &wg, timeout, progress, hits, ctx, hitCounter)
	}
	// extract the ip range
	var prefixes []netip.Prefix
	if input == "0" {
		rand.Shuffle(len(ips), func(i, j int) {
			ips[i], ips[j] = ips[j], ips[i]
		})
		for _, cidr := range ips {
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				log.Println(err)
				continue
			}
			prefixes = append(prefixes, prefix)
		}
	} else {
		prefix, err := netip.ParsePrefix(ips[selection-1])
		if err != nil {
			panic(err)
		}
		prefixes = append(prefixes, prefix)
	}
	for _, prefix := range prefixes {
		count := 1 << (32 - prefix.Bits())
		totalIpCount += uint64(count)
	}

	go report.UpdateProgress(&totalIpCount, progress, done, hitCounter, stop)
	go func() {
		defer close(jobs)
		for _, prefix := range prefixes {
			ip := prefix.Masked().Addr()
			count := 1 << (32 - prefix.Bits())
			for range count {
				select {
				case jobs <- ip.String():
					ip = ip.Next()
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	wg.Wait()

	close(hits)
	close(done)
	close(hitCounter)

	<-hitCounter
	<-doneHits

	if ctx.Err() != nil {
		fmt.Printf("Scan cancelled.\n")
	} else {
		log.Println("Scan completed.")
	}
}
