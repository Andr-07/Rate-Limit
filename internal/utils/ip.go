package utils

import (
	"fmt"
	"net"
	"net/http"
)

func GetClientIP(r *http.Request) (string, error) {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip, nil
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip, nil
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("invalid remote address: %v", err)
	}
	return host, nil
}
