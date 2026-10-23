# Go Cloudflare IP Scanner

[![Latest release](https://img.shields.io/github/v/release/proarash/go-cf-scanner?label=latest&style=flat-square)](https://github.com/proarash/go-cf-scanner/releases/latest)
[![Downloads](https://img.shields.io/github/downloads/proarash/go-cf-scanner/total?style=flat-square)](https://github.com/proarash/go-cf-scanner/releases)
[![Build](https://img.shields.io/github/actions/workflow/status/proarash/go-cf-scanner/go.yml?branch=main&style=flat-square)](https://github.com/proarash/go-cf-scanner/actions)

[![Windows](https://img.shields.io/badge/Windows-x64-0078D6?style=for-the-badge&logo=windows&logoColor=white)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-windows-x64.exe)
[![macOS](https://img.shields.io/badge/macOS-Apple%20Silicon-000000?style=for-the-badge&logo=apple&logoColor=white)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-macos-arm64)
[![Linux](https://img.shields.io/badge/Linux-x64-FCC624?style=for-the-badge&logo=linux&logoColor=black)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-linux-x64)
[![Android](https://img.shields.io/badge/Android-ARM64-3DDC84?style=for-the-badge&logo=android&logoColor=white)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-android-arm64)

This program scans Cloudflare's IPv4 ranges to find IP addresses that are reachable from **your** network. It is built for heavily filtered networks (such as Iran during internet blackouts), where only some Cloudflare IPs get through, and those IPs change depending on the ISP, the city and the day.

The IPs it finds can be used as the "clean IP" (the address you connect to) in proxy setups that run behind Cloudflare, such as v2Ray, Xray or a VLESS/VMess config with a CDN.

## Download

Each link always downloads the newest version.

| OS | Downloads |
| --- | --- |
| Windows | [x64](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-windows-x64.exe) · [ARM64](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-windows-arm64.exe) · [x86 (32-bit)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-windows-x86.exe) |
| macOS | [Apple Silicon (M1+)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-macos-arm64) · [Intel](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-macos-x64) |
| Linux | [x64](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-linux-x64) · [ARM64](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-linux-arm64) · [ARM (32-bit)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-linux-arm) · [x86 (32-bit)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-linux-x86) |
| Android | [ARM64 (Termux)](https://github.com/proarash/go-cf-scanner/releases/latest/download/cf-scanner-android-arm64) |

On Linux, macOS and Android, make the file executable before running it: `chmod +x cf-scanner-*`

## Build

```bash
go build -trimpath -ldflags="-s -w" -o main .
```

`-s -w` removes debug information to make the binary smaller. `-trimpath` removes the local folder paths of the build machine from the binary.

To build for every supported platform at once, run one of these from any folder:

```bash
sh scripts/build.sh            # Linux / macOS / Git Bash
```

```powershell
pwsh scripts/build.ps1         # Windows PowerShell
```

The binaries are written to `out/`, named `cf-scanner-<os>-<arch>`, for example `cf-scanner-windows-x64.exe`, `cf-scanner-linux-arm64` or `cf-scanner-macos-arm64`. The scripts build for Windows (x86, x64, arm64), Linux (x64, x86, arm64, 32-bit arm), macOS (x64, arm64) and Android (arm64). They also add `CGO_ENABLED=0` (no C dependencies, so the binaries run on any system), `-buildvcs=false` and `-buildid=` so the binaries are smaller and carry no build-machine details.

Android builds for x86, x64 and 32-bit arm are not included, because Go can only build them with the Android NDK (C toolchain) installed. On Android, run the `arm64` binary from a terminal app such as Termux.

## Quick start

```bash
./main                          # fast TCP scan (default)
./main -tls                     # also check the TLS handshake
./main -http -timeout 3s        # full check: TCP + TLS + HTTP request
```

The program shows Cloudflare's IP ranges and asks which one to scan:

```text
[0] Randomize and scan all
[1] 173.245.48.0/20
[2] 103.21.244.0/22
...
Select CIDR to scan:
```

- Enter a number to scan only that range.
- Enter `0` to scan all ranges in random order.

While it runs you see a live status line:

```text
[Progress]: 0.06% | [Count]: 986 | [Hits]: 6 | Press CTRL + C to stop
```

When it finishes, reaches the hit limit, or you press `CTRL + C`, it prints the fastest IPs it found:

```text
Fastest 3 hits (all saved to hits.txt):
  104.20.198.43   224ms
  172.67.229.204  234ms
  172.69.65.116   240ms
```

Every hit is also added to `hits.txt` as soon as it is found, one per line as `IP latency_in_ms`. The file is appended to and never cleared, so results from earlier runs stay in it.

## Flags

All flags are optional.

| Flag | Default | What it does |
| --- | --- | --- |
| `-bots` | `500` | Number of IPs tested at the same time. Higher is faster, but on weak or mobile connections too many connections at once can trigger rate limiting or drop your connection. Try `100`–`200` on mobile networks. |
| `-timeout` | `1s` | How long to wait for each IP before giving up. Accepts values like `800ms`, `2s` or `3s`. On slow or throttled connections, and with `-tls`/`-http`, use `2s`–`3s`, or working IPs will be counted as failures. |
| `-hits` | `10` | Stop after this many working IPs are found. `0` scans the whole selection. |
| `-tls` | off | Also require a successful TLS handshake (see "How it works"). |
| `-http` | off | Also require TLS and an HTTP `200` reply from `/cdn-cgi/trace`. Includes `-tls`. This is the strictest and most reliable check. |
| `-sni` | `cloudflare.com` | Hostname sent during the TLS handshake and in the HTTP `Host` header. Used only with `-tls` or `-http`. **Set it to the domain you actually use**, for example your own domain behind Cloudflare. |
| `-port` | `443` | Port to connect to. Cloudflare also serves HTTPS on `2053`, `2083`, `2087`, `2096` and `8443`. |
| `-top` | `10` | How many of the fastest hits to print at the end. |

Examples:

```bash
# scan one range with mobile-friendly settings
./main -bots 150 -timeout 2s

# find 20 IPs that really work for your domain
./main -http -sni example.com -timeout 3s -hits 20

# test every IP in the selection (no hit limit)
./main -hits 0
```

## How it works

### Why scan at all?

Cloudflare serves every website behind it from the same large pool of IP addresses. Any Cloudflare IP can deliver traffic for any Cloudflare domain. A filtered network usually can't block Cloudflare completely, because too many ordinary websites depend on it, so it blocks or slows down **parts** of Cloudflare's address space instead. Which parts are blocked differs between ISPs and changes over time. The only reliable way to find addresses that work for you is to test them from your own connection, which is what this program does.

### Step 1: Pick the addresses

The program has Cloudflare's official IPv4 ranges built in (the list at <https://www.cloudflare.com/ips-v4>), about 1.5 million addresses in total.

- **One range:** its IPs are tested in order.
- **Option `0`:** all ranges are scanned in a random order that visits every IP exactly once. This spreads the tests across the whole address space, so with a hit limit you get IPs from many different ranges instead of the first few addresses of one block. Using a random order also makes the scan look less like a sequential sweep.

The order is computed on the fly (`index → (a × index + c) mod total`, with a random `a` and `c`), so the program never has to keep a list of 1.5 million addresses in memory.

### Step 2: Test the addresses

A pool of workers (set by `-bots`) takes addresses from a queue and tests them at the same time. Each test has one of three levels:

| Level | Flag | What is tested | What it tells you |
| --- | --- | --- | --- |
| TCP | (default) | A TCP connection to the port is opened. | The IP is not completely blocked. Fast, but a weak signal: filtering often lets the TCP connection through and blocks the next step. |
| TLS | `-tls` | After TCP, a TLS (HTTPS) handshake is done using the `-sni` hostname. | The filter does not cut the encrypted connection for that hostname. Filters often block based on this hostname, which is why `-sni` should be your real domain. |
| HTTP | `-http` | After TLS, a request for `https://<sni>/cdn-cgi/trace` is sent and must return `200`. | Real data passes both ways through Cloudflare. This is closest to what your proxy will actually do. |

Each level adds about one network round trip, so stricter checks are slower per IP and need a longer `-timeout`. IPs that don't answer cost the same at every level: the full timeout.

**Suggested workflow:** start with a fast TCP scan to see which ranges respond at all. Then scan the best range again with `-http -sni yourdomain.com` to find the IPs that actually work.

### Step 3: Measure and save

For each working IP the program records the **latency**: how long the whole check took, from opening the connection to the end of the last step. A lower latency usually means a faster, more stable proxy connection. Hits are written to `hits.txt` immediately, so nothing is lost if the program is closed suddenly. At the end they are sorted and the fastest ones are printed.

### Stopping

The scan stops when one of these happens:

- the `-hits` limit is reached;
- every address in the selection has been tested;
- you press `CTRL + C`.

In every case the workers shut down cleanly and all results are saved.

## Project structure

```text
main.go                        flags, range selection, random ordering, worker pool
internal/cloudflare/ranges.go  Cloudflare IPv4 range list
internal/scanner/scanner.go    worker loop: takes IPs from the queue and tests them
internal/scanner/tcping.go     the TCP / TLS / HTTP check and latency measurement
internal/report/hits.go        writes hits to hits.txt and sorts them by latency
internal/report/progress.go    live progress line
```

## Ideas for future improvements

- Test several ports per IP (`443`, `8443`, `2053`, `2083`, `2087`, `2096`) and report which ones are open.
- Repeat the check several times per IP and report average latency, jitter and failure rate, so unstable IPs can be filtered out.
- Download and upload speed tests through the best IPs.
- WebSocket upgrade test, for v2Ray/Xray WebSocket transport.
- Try several SNI values to see which hostnames are filtered.
- IPv6 range support.
- Resume an interrupted scan, and export results to CSV/JSON.
