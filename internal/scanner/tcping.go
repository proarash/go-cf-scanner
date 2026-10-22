package scanner

import (
	"context"
	"net"
	"strconv"
	"time"
)

func GetTcping(ipAddr string, port, timeout int, ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	addr := net.JoinHostPort(ipAddr, strconv.Itoa(port))
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp4", addr)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
