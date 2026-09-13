package subnet

import (
	"log/slog"
	"net"
	"net/http"
	"os"
)

func Middleware(trustedSubnet string) func(http.Handler) http.Handler {
	_, trustedNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		slog.Error("cidr configuration error", "err", err)
		os.Exit(1)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedNet == nil {
				http.Error(w, "Internal Server Error: Invalid network configuration", http.StatusInternalServerError)
				return
			}

			realIPStr := r.Header.Get("X-Real-IP")

			if realIPStr == "" {
				http.Error(w, "Forbidden: Missing X-Real-IP header", http.StatusForbidden)
				return
			}

			realIP := net.ParseIP(realIPStr)
			if realIP == nil {
				http.Error(w, "Forbidden: Invalid IP format", http.StatusForbidden)
				return
			}

			if !trustedNet.Contains(realIP) {
				slog.Warn("Access forbidden", "ip", realIPStr, "subnet", trustedSubnet)
				http.Error(w, "Forbidden: IP address is not trusted", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
