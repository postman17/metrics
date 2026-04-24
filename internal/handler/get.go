package handler

import (
	"fmt"
	"net/http"

	mem "github.com/postman17/metrics/internal/repository"
)

func GetMetricValuePage(memory *mem.MemStorage) http.HandlerFunc {
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
		value := memory.GetTypeValue(metricName)
		if value == nil {
			rw.WriteHeader(http.StatusNotFound)
			return
		}
		if metricType == "gauge" {
			rw.Write([]byte(fmt.Sprintf("%g", value)))
		} else {
			rw.Write([]byte(fmt.Sprintf("%d", value)))
		}
		rw.Header().Set("Content-Type", "text/plain")
		rw.WriteHeader(http.StatusOK)
		fmt.Println(memory)
	}
}
