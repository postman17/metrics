// Package handler реализует HTTP-обработчики сервера мониторинга:
// обновление и получение метрик в форматах plain-text и JSON,
// главная страница со списком метрик, а также проверка связи с БД.
package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	models "github.com/postman17/metrics/internal/model"
	mem "github.com/postman17/metrics/internal/repository"
)

// GetMetricValuePage возвращает HTTP-обработчик, который отдаёт значение метрики
// в текстовом виде по пути /value/{type}/{name}.
func GetMetricValuePage(storage mem.MetricsRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		metricType := r.PathValue("type")
		metricName := r.PathValue("name")
		if metricType == "" || metricName == "" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		if metricType != "gauge" && metricType != "counter" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		value := storage.GetTypeValue(metricName)
		if value == nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		if metricType == "gauge" {
			_, _ = rw.Write([]byte(fmt.Sprintf("%g", value)))
		} else {
			_, _ = rw.Write([]byte(fmt.Sprintf("%d", value)))
		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		slog.Debug("get metric value page", "storage", storage)
	}
}

// GetMetricValue возвращает HTTP-обработчик, который отдаёт значение метрики
// в JSON-формате по пути /value/. Ожидает POST-запрос с JSON-телом GetMetricRequest.
func GetMetricValue(storage mem.MetricsRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req models.GetMetricRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metricType := req.MType
		metricName := req.ID
		if metricType == "" || metricName == "" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		if metricType != "gauge" && metricType != "counter" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		value := storage.GetTypeValue(metricName)
		if value == nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		resp := models.Metrics{
			ID:    metricName,
			MType: metricType,
		}
		if metricType == "gauge" {
			if val, ok := value.(float64); ok {
				resp.Value = &val
			}

		} else {
			if val, ok := value.(int64); ok {
				resp.Delta = &val
			}
		}

		rw.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(rw)
		if err := enc.Encode(resp); err != nil {
			slog.Info("error encoding response")
			return
		}
	}
}
