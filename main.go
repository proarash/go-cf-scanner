package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net/netip"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/proarash/go-cf-scanner/internal/cloudflare"
	"github.com/proarash/go-cf-scanner/internal/report"
	"github.com/proarash/go-cf-scanner/internal/scanner"
	"golang.org/x/term"
)

func main() {
	bots := flag.Int("bots", 500, "number of concurrent workers")
	timeout := flag.Duration("timeout", time.Second, "per-IP timeout (e.g. 800ms, 2s)")
	maxHits := flag.Uint64("hits", 10, "stop after this many hits (0 = scan everything)")
	useTLS := flag.Bool("tls", false, "also require a successful TLS handshake (default: TCP connect only)")
	useHTTP := flag.Bool("http", false, "also require TLS + HTTP 200 from /cdn-cgi/trace (implies -tls)")
	sni := flag.String("sni", "cloudflare.com", "SNI / Host used by -tls and -http")
	port := flag.Int("port", 443, "port to probe")
	top := flag.Int("top", 10, "number of fastest hits to print at the end")
	flag.Parse()

	cfg := scanner.Config{Port: *port, Timeout: *timeout, Mode: scanner.ModeTCP, SNI: *sni}
	switch {
	case *useHTTP:
		cfg.Mode = scanner.ModeHTTP
	case *useTLS:
		cfg.Mode = scanner.ModeTLS
	}
	if *bots < 1 {
		log.Fatalf("-bots must be at least 1")
	}

	ips := cloudflare.Ip4Ranges()
	fmt.Println("[0] Randomize and scan all")
	for index, ip := range ips {
		fmt.Printf("[%v] %v\n", index+1, ip)
	}
	fmt.Printf("Select CIDR to scan: ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	input = strings.TrimSpace(input)
	if err != nil && input == "" {
		log.Fatalf("failed to read selection: %v", err)
	}
	selection, err := strconv.Atoi(input)
	if err != nil || selection < 0 || selection > len(ips) {
		log.Fatalf("invalid selection %q, expected a number between 0 and %d", input, len(ips))
	}

	var prefixes []netip.Prefix
	if selection == 0 {
		log.Println("Selected 0, scanning all ranges in random order.")
		for _, cidr := range ips {
			prefix, err := netip.ParsePrefix(cidr)
			if err != nil {
				log.Println(err)
				continue
			}
			prefixes = append(prefixes, prefix.Masked())
		}
	} else {
		log.Printf("Selected %v | %v\n", selection, ips[selection-1])
		prefix, err := netip.ParsePrefix(ips[selection-1])
		if err != nil {
			log.Fatal(err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}

	// offsets[i] is the global index of the first IP in prefixes[i]
	offsets := make([]uint64, len(prefixes))
	var total uint64
	for i, prefix := range prefixes {
		offsets[i] = total
		total += 1 << (32 - prefix.Bits())
	}

	log.Printf("Bots: %v | Timeout: %v | Mode: %v | Total IPs: %v\n", *bots, *timeout, cfg.Mode, total)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	jobs := make(chan string, *bots*2)
	hits := make(chan scanner.Hit, *bots)
	allHits := make(chan []scanner.Hit, 1)
	progressDone := make(chan struct{})
	var progress, hitCount atomic.Uint64
	var limitReached atomic.Bool

	go report.SaveHits(hits, func() {
		if n := hitCount.Add(1); *maxHits > 0 && n >= *maxHits && !limitReached.Swap(true) {
			stop()
		}
	}, allHits)
	progressExited := make(chan struct{})
	go func() {
		defer close(progressExited)
		report.UpdateProgress(total, &progress, &hitCount, progressDone)
	}()

	var wg sync.WaitGroup
	for range *bots {
		wg.Add(1)
		go scanner.ScanWorker(ctx, &wg, cfg, jobs, &progress, hits)
	}

	go func() {
		defer close(jobs)
		// Visit every index exactly once in pseudo-random order: i -> (a*i + c) mod total, gcd(a, total) == 1.
		a, c := uint64(1), uint64(0)
		if selection == 0 {
			for {
				a = rand.Uint64N(total-1) + 1
				if gcd(a, total) == 1 {
					break
				}
			}
			c = rand.Uint64N(total)
		}
		for i := range total {
			idx := (mulmod(a, i, total) + c) % total
			p := len(offsets) - 1
			for offsets[p] > idx {
				p--
			}
			addr := prefixes[p].Addr().As4()
			n := uint32(addr[0])<<24 | uint32(addr[1])<<16 | uint32(addr[2])<<8 | uint32(addr[3])
			n += uint32(idx - offsets[p])
			ip := netip.AddrFrom4([4]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
			select {
			case jobs <- ip.String():
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	close(hits)
	found := <-allHits
	close(progressDone)
	<-progressExited
	defer waitForKey()

	switch {
	case limitReached.Load():
		log.Printf("Reached %d hits, stopping.", *maxHits)
	case ctx.Err() != nil:
		log.Println("Scan cancelled.")
	default:
		log.Println("Scan completed.")
	}

	if len(found) == 0 {
		fmt.Println("\nNo working IPs found. Try another range, a longer -timeout or fewer -bots.")
		return
	}

	best := found[:min(*top, len(found))]
	fmt.Printf("\nFastest %d of %d working IPs:\n", len(best), len(found))
	for _, hit := range best {
		fmt.Printf("  %-15s %v\n", hit.IP, hit.Latency.Round(time.Millisecond))
	}

	fmt.Println("\nClean IPs (ready to copy):")
	fmt.Println()
	for _, hit := range best {
		fmt.Println(hit.IP)
	}

	hitsPath, err := filepath.Abs(report.HitsFile)
	if err != nil {
		hitsPath = report.HitsFile
	}
	fmt.Printf("\nAll %d working IPs from this scan were saved to:\n  %s\n", len(found), hitsPath)
}

// waitForKey keeps the window open until a key is pressed, so results stay visible when the
// program was started by double-clicking. It does nothing when stdin is not a terminal.
func waitForKey() {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return
	}
	fmt.Print("\nPress any key to exit...")
	state, err := term.MakeRaw(fd)
	if err != nil {
		return
	}
	defer term.Restore(fd, state)
	os.Stdin.Read(make([]byte, 1))
	fmt.Print("\r\n")
}

func gcd(a, b uint64) uint64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// mulmod is safe from overflow because total IP count is well below 2^32.
func mulmod(a, b, m uint64) uint64 {
	return (a % m) * (b % m) % m
}
