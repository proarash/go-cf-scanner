# Cloudflare IP Scanner for v2Ray – Project Workflow

## Phase 1: Data Acquisition
1. Fetch official Cloudflare IPv4 ranges from `https://www.cloudflare.com/ips-v4`
2. Fetch official Cloudflare IPv6 ranges from `https://www.cloudflare.com/ips-v6` (optional)
3. Parse and store CIDR blocks in memory
4. Expand each CIDR into individual IP addresses (or generate on-the-fly to save memory)

## Phase 2: Pre-Scan Filtering
5. Remove duplicate IPs
6. Optionally filter out specific subnets that are known to be slow or blocked
7. Randomize the IP list order to avoid pattern detection
8. Split IPs into chunks for concurrent scanning

## Phase 3: Ping/Liveness Test (First Pass)
9. For each IP, send ICMP echo request (ping) with 1-2 second timeout
10. Send 3 pings per IP to calculate average and loss percentage
11. Record: min RTT, max RTT, average RTT, packet loss %
12. Discard IPs with loss > 20% or average RTT > 300ms
13. If ICMP is blocked, fall back to TCP ping (SYN to port 443)

## Phase 4: TCP Port Scan
14. For surviving IPs, test common Cloudflare ports: 443, 8443, 2053, 2083, 2087, 2096
15. Attempt TCP connection with 1-second timeout
16. Record which ports are open
17. Keep only IPs with at least one open port
18. Measure TCP handshake time (RTT) for each open port

## Phase 5: TLS & HTTP Validation
19. For IPs with port 443 or other HTTPS ports open, perform TLS handshake
20. Verify certificate is from Cloudflare (CN or SAN contains cloudflare.com)
21. Send HTTP GET request with a valid Host header (your domain proxied by Cloudflare)
22. Check response headers for Cloudflare signatures: `cf-ray`, `Server: cloudflare`
23. Measure time to first byte (TTFB)
24. Discard IPs that don't return Cloudflare headers or timeout

## Phase 6: Latency Benchmark
25. For validated IPs, send 10 sequential HTTPS requests to same endpoint
26. Calculate average, median, and standard deviation of response times
27. Sort IPs by average latency (lowest first)
28. Discard IPs with unstable latency (high standard deviation > 30ms)

## Phase 7: Download Speed Test
29. For top 100-500 IPs (from latency sort), test download speed
30. Download a 1MB to 10MB test file hosted behind Cloudflare
31. Measure total download time with 5-10 second timeout
32. Calculate speed in Mbps: `(file_size_mb * 8) / download_time_sec`
33. Repeat 3 times per IP and take average
34. Record download speed for each IP

## Phase 8: Upload Speed Test
35. For same IPs, test upload speed
36. Generate random payload of 500KB to 5MB
37. Send HTTP POST request to same endpoint
38. Measure total upload time with timeout
39. Calculate upload speed in Mbps
40. Repeat 3 times and take average

## Phase 9: Scoring & Ranking
41. Assign weights to metrics (example):
    - Ping (30% weight, lower is better)
    - Download speed (40% weight, higher is better)
    - Upload speed (30% weight, higher is better)
42. Calculate composite score for each IP
43. Rank IPs by score (highest first)
44. Generate top 20-50 "best" IPs

## Phase 10: v2Ray Compatibility Test
45. Take top 10 IPs from ranking
46. Replace IP in your v2Ray config with each candidate
47. Run v2Ray with each config for 2-5 minutes
48. Measure real proxy throughput and stability
49. Check for connection drops, timeouts, or throttling
50. Output final validated IP list with performance metrics

## Phase 11: Output & Reporting
51. Export results to JSON/CSV format
52. Generate human-readable markdown report with:
    - Top 10 IPs with scores
    - Latency distribution chart (optional)
    - Speed test results table
53. Save raw scan data for future comparison
54. Create simple `good-ips.txt` with format `IP:PORT` for quick copy-paste

## Phase 12: Safety & Rate Limiting
55. Add random delays between requests (50-200ms)
56. Limit concurrent workers to 50-100
57. Add jitter to scan timing to avoid patterns
58. Implement exponential backoff on rate-limit errors (HTTP 429)
59. Stop scanning if Cloudflare returns CAPTCHA or 403 block page

## Phase 13: Optional Enhancements
60. Add IPv6 scanning support
61. Implement resume functionality (save progress)
62. Test different SNI values (different Host headers)
63. Add WebSocket upgrade test (for v2Ray WS transport)
64. Schedule periodic rescans (daily/weekly)
65. Compare results over time to detect IP degradation
66. Add notification system (Telegram/email) when new fast IP found