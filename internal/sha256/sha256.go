package sha256

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

func getHash(key string) string {
	hash := sha256.New()
	hash.Write([]byte(key))
	return hex.EncodeToString(hash.Sum(nil))
}

func Middleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if key == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientHash := r.Header.Get("HashSHA256")
			hash := getHash(key)

			if clientHash != "" && clientHash != hash {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("HashSHA256", hash)
			next.ServeHTTP(w, r)
		})
	}
}
