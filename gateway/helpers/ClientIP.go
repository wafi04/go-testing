package helpers

import (
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	remoteAddr := r.RemoteAddr
	if strings.Contains(remoteAddr, ":") {
		return strings.Split(remoteAddr, ":")[0]
	}

	return remoteAddr
}
