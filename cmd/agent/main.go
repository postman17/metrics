package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
)

var sendGzipJSONRetryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

func getHash(key string) string {
	hash := sha256.New()
	hash.Write([]byte(key))
	return hex.EncodeToString(hash.Sum(nil))
}

func sendGzipJSON(client http.Client, url string, jsonData []byte, key string) (*http.Response, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write(jsonData); err != nil {
		slog.Error("Gzip error", "err", err)
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		slog.Error("Error gzip compress", "err", err)
		return nil, err
	}

	body := buf.Bytes()
	var lastErr error
	for attempt := 0; attempt <= len(sendGzipJSONRetryDelays); attempt++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			slog.Error("Create request error", "err", err)
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("HashSHA256", getHash(key))

		resp, err := client.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if attempt == len(sendGzipJSONRetryDelays) {
			break
		}

		delay := sendGzipJSONRetryDelays[attempt]
		slog.Warn("Send request error, retrying", "err", err, "attempt", attempt+1, "delay", delay)
		time.Sleep(delay)
	}

	fmt.Printf("Send request error: %v\n", lastErr)
	return nil, lastErr
}

func SendBatchRequest(client http.Client, config Config, metrics models.MetricsList) (*http.Response, error) {
	url := fmt.Sprintf("%s/updates/", config.RunAddr)

	jsonData, err := easyjson.Marshal(metrics)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return nil, err
	}

	resp, err := sendGzipJSON(client, url, jsonData, config.Key)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}

func gaugeMetric(name string, value float64) models.Metrics {
	v := value
	return models.Metrics{
		ID:    name,
		MType: "gauge",
		Value: &v,
	}
}

func counterMetric(name string, delta int64) models.Metrics {
	d := delta
	return models.Metrics{
		ID:    name,
		MType: "counter",
		Delta: &d,
	}
}

func runtimeMetrics(m *runtime.MemStats) models.MetricsList {
	return models.MetricsList{
		gaugeMetric("Alloc", float64(m.Alloc)),
		gaugeMetric("BuckHashSys", float64(m.BuckHashSys)),
		gaugeMetric("Frees", float64(m.Frees)),
		gaugeMetric("GCCPUFraction", float64(m.GCCPUFraction)),
		gaugeMetric("GCSys", float64(m.GCSys)),
		gaugeMetric("HeapAlloc", float64(m.HeapAlloc)),
		gaugeMetric("HeapIdle", float64(m.HeapIdle)),
		gaugeMetric("HeapInuse", float64(m.HeapInuse)),
		gaugeMetric("HeapObjects", float64(m.HeapObjects)),
		gaugeMetric("HeapReleased", float64(m.HeapReleased)),
		gaugeMetric("HeapSys", float64(m.HeapSys)),
		gaugeMetric("LastGC", float64(m.LastGC)),
		gaugeMetric("Lookups", float64(m.Lookups)),
		gaugeMetric("MCacheInuse", float64(m.MCacheInuse)),
		gaugeMetric("MCacheSys", float64(m.MCacheSys)),
		gaugeMetric("MSpanInuse", float64(m.MSpanInuse)),
		gaugeMetric("MSpanSys", float64(m.MSpanSys)),
		gaugeMetric("Mallocs", float64(m.Mallocs)),
		gaugeMetric("NextGC", float64(m.NextGC)),
		gaugeMetric("NumForcedGC", float64(m.NumForcedGC)),
		gaugeMetric("NumGC", float64(m.NumGC)),
		gaugeMetric("OtherSys", float64(m.OtherSys)),
		gaugeMetric("PauseTotalNs", float64(m.PauseTotalNs)),
		gaugeMetric("StackInuse", float64(m.StackInuse)),
		gaugeMetric("StackSys", float64(m.StackSys)),
		gaugeMetric("Sys", float64(m.Sys)),
		gaugeMetric("TotalAlloc", float64(m.TotalAlloc)),
		gaugeMetric("RandomValue", rand.Float64()),
		counterMetric("PollCount", 1),
	}
}

func main() {
	config := parseFlags()

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	var m runtime.MemStats
	for {
		select {
		case t1 := <-tickerPoll.C:
			runtime.ReadMemStats(&m)
			slog.Info("Fast tic:", "time", t1)
		case t2 := <-tickerReport.C:
			_, err := SendBatchRequest(*client, config, runtimeMetrics(&m))
			if err != nil {
				slog.Error("Send batch metrics error", "err", err)
			}
			slog.Info("Slow tic:", "time", t2)
		}
	}
}
