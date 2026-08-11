package handler

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/mailru/easyjson"
	audit "github.com/postman17/metrics/internal/audit"
	models "github.com/postman17/metrics/internal/model"
	mem "github.com/postman17/metrics/internal/repository"
)

// updatesBufPool — пул буферов bytes.Buffer для чтения тела запроса
// в обработчике UpdatesMetric, уменьшает allocations.
var updatesBufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// emptyJSON используется как тело ответа при успешном пакетном обновлении.
var emptyJSON = []byte("{}")

// UpdatesMetric возвращает HTTP-обработчик для пакетного обновления метрик
// по пути /updates/. Ожидает POST-запрос с JSON-массивом MetricsList.
// После записи в хранилище уведомляет подписчиков аудита (pub).
func UpdatesMetric(storage mem.MetricsRepository, pub *audit.Pub) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		buf := updatesBufPool.Get().(*bytes.Buffer)
		buf.Reset()
		_, err := io.Copy(buf, r.Body)
		_ = r.Body.Close()
		if err != nil {
			updatesBufPool.Put(buf)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		var req models.MetricsList
		if unmarshalErr := easyjson.Unmarshal(buf.Bytes(), &req); unmarshalErr != nil {
			updatesBufPool.Put(buf)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		updatesBufPool.Put(buf)

		err = storage.AddBatch(req)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metrics := make([]string, 0, len(req))
		for _, value := range req {
			metrics = append(metrics, value.ID)
		}
		pub.Notify(r, metrics)

		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(emptyJSON)
		slog.Debug("update metric", "storage", storage)
	}
}
