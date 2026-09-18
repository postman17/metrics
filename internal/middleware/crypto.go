package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"log/slog"
	"net/http"

	rsacrypto "github.com/postman17/metrics/internal/rsacrypto"
)

// Crypto возвращает HTTP-middleware, который расшифровывает тело запроса
// (application/octet-stream) приватным RSA-ключом и передает дальше уже
// открытый JSON. Если ключ не задан или тип контента иной, запрос проходит
// без изменений.
func Crypto(privKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Content-Type") != "application/octet-stream" || privKey == nil {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			_ = r.Body.Close()
			if err != nil {
				slog.Error("crypto middleware: read body", "err", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			decrypted, err := rsacrypto.Decrypt(privKey, body)
			if err != nil {
				slog.Error("crypto middleware: decrypt", "err", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.Header.Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}
