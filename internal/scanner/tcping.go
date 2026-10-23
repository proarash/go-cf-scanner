package scanner

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Mode string

const (
	ModeTCP  Mode = "tcp"
	ModeTLS  Mode = "tls"
	ModeHTTP Mode = "http"
)

var dialer net.Dialer

// Probe connects to ipAddr:port and returns the time taken by the check selected by mode.
// tcp: TCP connect only. tls: TCP + TLS handshake with sni. http: TLS + GET /cdn-cgi/trace.
func Probe(ctx context.Context, ipAddr string, port int, timeout time.Duration, mode Mode, sni string) (time.Duration, bool) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp4", net.JoinHostPort(ipAddr, strconv.Itoa(port)))
	if err != nil {
		return 0, false
	}
	defer conn.Close()
	if mode == ModeTCP {
		return time.Since(start), true
	}

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}
	tlsConn := tls.Client(conn, &tls.Config{ServerName: sni, NextProtos: []string{"http/1.1"}})
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return 0, false
	}
	if mode == ModeTLS {
		return time.Since(start), true
	}

	if _, err := fmt.Fprintf(tlsConn, "GET /cdn-cgi/trace HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", sni); err != nil {
		return 0, false
	}
	resp, err := http.ReadResponse(bufio.NewReader(tlsConn), nil)
	if err != nil {
		return 0, false
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, false
	}
	return time.Since(start), true
}
