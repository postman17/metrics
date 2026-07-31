package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/mailru/easyjson"
	audit "github.com/postman17/metrics/internal/audit"
	models "github.com/postman17/metrics/internal/model"
	mem "github.com/postman17/metrics/internal/repository"
)

func UpdatesMetric(storage mem.MetricsRepository, pub audit.Pub) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var req models.MetricsList
		if err := easyjson.Unmarshal(body, &req); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		err = storage.AddBatch(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		var metrics []string
		for _, value := range req {
			metrics = append(metrics, value.ID)
		}
		pub.Notify(r, metrics)

		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("{}"))
		slog.Debug("update metric", "storage", storage)
	}
}
