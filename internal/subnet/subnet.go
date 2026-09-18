package subnet

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

// Middleware проверяет, что IP-адрес агента из заголовка X-Real-IP принадлежит
// доверенной подсети. CIDR валидируется один раз при создании middleware, чтобы
// вызывающий код мог перехватить ошибку конфигурации на старте.
func Middleware(trustedSubnet string) (func(http.Handler) http.Handler, error) {
	_, trustedNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("parse trusted subnet %q: %w", trustedSubnet, err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}, nil
}
