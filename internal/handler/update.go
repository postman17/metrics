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

		if !memory.StoreNotSync {
			if err := memory.SaveToFile(); err != nil {
				slog.Error(
					"memory save to file failed", "err", err,
				)
			}
		}

		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		fmt.Println(memory)
	}
}

func UpdateMetric(memory *mem.MemStorage) http.HandlerFunc {
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
			memory.AddGauge(metricName, *metricGaugeValue)
		case "counter":
			if metricCounterValue == nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			memory.AddCounter(metricName, *metricCounterValue)
		}

		if !memory.StoreNotSync {
			if err := memory.SaveToFile(); err != nil {
				slog.Error(
					"memory save to file failed", "err", err,
				)
			}
		}

		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte("{}"))
		fmt.Println(memory)
	}
}
