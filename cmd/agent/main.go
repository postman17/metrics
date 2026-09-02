// Агент собирает runtime-метрики Go (runtime.MemStats) и периодически
// отправляет их на сервер мониторинга пакетными POST-запросами.
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/mailru/easyjson"
	cryptopkg "github.com/postman17/metrics/internal/crypto"
	models "github.com/postman17/metrics/internal/model"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// sendGzipJSONRetryDelays — задержки между повторными попытками отправки HTTP-запроса.
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

func loadPublicKey(path string) (*rsa.PublicKey, error) {
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file error: %w", err)
	}
	return cryptopkg.ParsePublicKey(pemBytes)
}

func sendGzipJSON(
	client http.Client,
	url string,
	jsonData []byte,
	key string,
	pubKey *rsa.PublicKey,
) (*http.Response, error) {
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

	var body []byte
	contentType := "application/json"

	if pubKey != nil {
		encData, err := cryptopkg.Encrypt(pubKey, buf.Bytes())
		if err != nil {
			slog.Error("Encryption error", "err", err)
			return nil, err
		}
		body = encData
		contentType = "application/octet-stream"
	} else {
		body = buf.Bytes()
	}

	var lastErr error
	for attempt := 0; attempt <= len(sendGzipJSONRetryDelays); attempt++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			slog.Error("Create request error", "err", err)
			return nil, err
		}

		req.Header.Set("Content-Type", contentType)
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

// SendBatchRequest отправляет срез метрик на сервер пакетным POST-запросом
// по пути /updates/ с gzip-сжатием тела и подписью HashSHA256.
func SendBatchRequest(
	client http.Client,
	config Config,
	metrics models.MetricsList,
	pubKey *rsa.PublicKey,
) (*http.Response, error) {
	url := fmt.Sprintf("%s/updates/", config.RunAddr)

	jsonData, err := easyjson.Marshal(metrics)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return nil, err
	}

	resp, err := sendGzipJSON(client, url, jsonData, config.Key, pubKey)
	if err != nil {
		return nil, err
	}
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

func sendData(
	client *http.Client,
	config Config,
	m *runtime.MemStats,
	pubKey *rsa.PublicKey,
) {
	resp, err := SendBatchRequest(*client, config, runtimeMetrics(m), pubKey)
	if err != nil {
		slog.Error("Send batch metrics error", "err", err)
	} else {
		_ = resp.Body.Close()
	}
}

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	fmt.Println("Build version:", buildVersion)
	if buildDate == "" {
		buildDate = "N/A"
	}
	fmt.Println("Build date:", buildDate)
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Println("Build commit:", buildCommit)

	config := parseFlags()
	var pubKey *rsa.PublicKey
	if config.CryptoKey != "" {
		pubKeyLoad, err := loadPublicKey(config.CryptoKey)
		if err != nil {
			fmt.Printf("Load pem error: %v\n", err)
		}
		pubKey = pubKeyLoad
	}

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	var m runtime.MemStats
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	for {
		select {
		case t1 := <-tickerPoll.C:
			runtime.ReadMemStats(&m)
			slog.Info("Fast tic:", "time", t1)
		case t2 := <-tickerReport.C:
			sendData(
				client,
				config,
				&m,
				pubKey,
			)
			slog.Info("Slow tic:", "time", t2)
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			done := make(chan struct{})
			go func() {
				sendData(client, config, &m, pubKey)
				close(done)
			}()
			select {
			case <-done:
				slog.Info("Final data sent successfully. Agent done.")
			case <-shutdownCtx.Done():
				slog.Error("Shutdown timed out!")
			}
			cancel()
			slog.Info("Agent done")
			return
		}
	}
}
