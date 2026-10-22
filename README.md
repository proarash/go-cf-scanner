# Go Cloudflare IP Scanner

This program scans Cloudflare CIDR ranges to find responsive IP addresses based on your ISP's network rules and connectivity.

## Build

Build a smaller stripped binary:

```bash
go build -trimpath -ldflags="-s -w" -o main .
```

The `-s -w` linker flags remove symbol tables and debug information, reducing the binary size. `-trimpath` strips local filesystem paths from the binary, so recorded paths become module-relative (`github.com/proarash/go-cf-scanner/main.go`) instead of absolute build-machine paths.

## Run

The program takes no arguments or flags. Run it and pick a range interactively:

```bash
./main
```

It prints the Cloudflare IPv4 ranges and prompts for a selection:

```text
[0] Randomize and scan all
[1] 173.245.48.0/20
[2] 103.21.244.0/22
...
Select CIDR to scan:
```

- Enter the number of a single CIDR to scan only that range.
- Enter `0` to shuffle every Cloudflare range and scan all of them.

Scan settings are currently fixed in [main.go](main.go):

```text
bots:    100 concurrent workers
timeout: 3 seconds per connection
method:  TCP connect on port 443
```

Press `CTRL + C` at any time to stop the scan; the program shuts the workers down cleanly and flushes results to disk.

## How It Works

The program loads the embedded Cloudflare CIDR list, expands the selected prefixes into individual IP addresses, and feeds them to a pool of concurrent workers over a job channel.

Each worker attempts a TCP connection to port `443` and waits until the configured timeout expires. Responsive IP addresses are appended to `hits.txt`, and a live progress line reports the percentage scanned.

## Project Structure

```text
main.go                        entry point: CIDR selection, worker pool, IP feeding
internal/cloudflare/ranges.go  Cloudflare IPv4 CIDR list
internal/scanner/scanner.go    worker loop consuming jobs
internal/scanner/tcping.go     TCP connect check on a given IP/port
internal/report/hits.go        buffered writer appending results to hits.txt
internal/report/progress.go    live progress ticker
```
