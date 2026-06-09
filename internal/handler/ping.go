package handler

import (
	"database/sql"
	"log/slog"
	"net/http"
)

func Ping(DB *sql.DB) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		err := DB.Ping()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			slog.Error("failed to ping db", "err", err)
			return
		}
		rw.WriteHeader(http.StatusOK)
	}
}
