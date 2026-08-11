package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
	mem "github.com/postman17/metrics/internal/repository"
)

// UpdateMetricPage возвращает HTTP-обработчик, который обновляет метрику
// по пути /update/{type}/{name}/{value} (plain-text формат).
func UpdateMetricPage(storage mem.MetricsRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		metricType := r.PathValue("type")
		metricName := r.PathValue("name")
		metricValue := r.PathValue("value")
		if metricType == "" || metricName == "" || metricValue == "" {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Println(metricType, metricName, metricValue)
			return
		}

		if metricType != "gauge" && metricType != "counter" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		if metricType == "gauge" {
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if storage.CheckCounterType(metricName) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddGauge(metricName, value)
		} else if metricType == "counter" {
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if storage.CheckGaugeType(metricName) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddCounter(metricName, value)
		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		slog.Debug("update metric page", "storage", storage)
	}
}

// UpdateMetric возвращает HTTP-обработчик, который обновляет одну метрику
// по пути /update/. Ожидает POST-запрос с JSON-телом Metrics.
func UpdateMetric(storage mem.MetricsRepository) http.HandlerFunc {
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
		defer func() { _ = r.Body.Close() }()

		var req models.Metrics
		if err := easyjson.Unmarshal(body, &req); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		metricType := req.MType
		metricName := req.ID
		metricCounterValue := req.Delta
		metricGaugeValue := req.Value
		if metricType == "" || metricName == "" {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		if metricType != "gauge" && metricType != "counter" {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		switch metricType {
		case "gauge":
			if metricGaugeValue == nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddGauge(metricName, *metricGaugeValue)
		case "counter":
			if metricCounterValue == nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			storage.AddCounter(metricName, *metricCounterValue)
		}

		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("{}"))
		slog.Debug("update metric", "storage", storage)
	}
}
