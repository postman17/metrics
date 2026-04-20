package handler

import (
	"net/http"
	"strconv"

	mem "github.com/postman17/metrics/internal/repository"
)

func UpdateMetricPage(memory *mem.MemStorage) http.HandlerFunc {
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
			if memory.CheckCounterType(metricName) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			memory.AddGauge(metricName, value)
		} else if metricType == "counter" {
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if memory.CheckGaugeType(metricName) {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			memory.AddCounter(metricName, value)
		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
	}
}
