// Агент собирает runtime-метрики Go (runtime.MemStats) и периодически
// отправляет их на сервер мониторинга. Поддерживается два транспорта:
// HTTP (пакетная отправка gzip+JSON) и gRPC (метод UpdateMetrics).
// Транспорт gRPC включается флагом -g / env GRPC.
package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/mailru/easyjson"
	cryptopkg "github.com/postman17/metrics/internal/crypto"
	models "github.com/postman17/metrics/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/postman17/metrics/pkg/proto"
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

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, address := range addrs {
		// Проверяем, что это IP-сеть и адрес не является loopback (127.0.0.1)
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("local ip address not found")
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

	localIP, err := getLocalIP()
	if err != nil {
		localIP = "127.0.0.1"
	}

	var lastErr error
	for attempt := 0; attempt <= len(sendGzipJSONRetryDelays); attempt++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			slog.Error("Create request error", "err", err)
			return nil, err
		}

		req.Header.Set("X-Real-IP", localIP)
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("HashSHA256", getHash(key))

		resp, err := client.Do(req)
		if err == nil && resp != nil {
			return resp, nil
		}

		if err == nil {
			err = errors.New("nil response")
		}
		lastErr = err
		if attempt == len(sendGzipJSONRetryDelays) {
			break
		}

		delay := sendGzipJSONRetryDelays[attempt]
		slog.Warn("Send request error, retrying", "err", err, "attempt", attempt+1, "delay", delay)
		time.Sleep(delay)
	}

	if lastErr == nil {
		lastErr = errors.New("send request failed")
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

func SendGRPCRequest(
	ctx context.Context,
	grpcClient pb.MetricsClient,
	metrics models.MetricsList,
) error {
	localIP, err := getLocalIP()
	if err != nil {
		localIP = "127.0.0.1"
	}

	req := &pb.UpdateMetricsRequest{
		Metrics: modelsToProto(metrics),
	}

	ctxWithIP := metadata.AppendToOutgoingContext(ctx, "x-real-ip", localIP)

	_, err = grpcClient.UpdateMetrics(ctxWithIP, req)
	if err != nil {
		return fmt.Errorf("grpc UpdateMetrics: %w", err)
	}
	return nil
}

func modelsToProto(metrics models.MetricsList) []*pb.Metric {
	result := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		pm := &pb.Metric{Id: m.ID}
		switch m.MType {
		case models.Counter:
			pm.Type = pb.Metric_COUNTER
			if m.Delta != nil {
				pm.Delta = *m.Delta
			}
		default:
			pm.Type = pb.Metric_GAUGE
			if m.Value != nil {
				pm.Value = *m.Value
			}
		}
		result = append(result, pm)
	}
	return result
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
	ctx context.Context,
	client *http.Client,
	grpcClient pb.MetricsClient,
	config Config,
	m *runtime.MemStats,
	pubKey *rsa.PublicKey,
) {
	metrics := runtimeMetrics(m)

	if grpcClient != nil {
		if err := SendGRPCRequest(ctx, grpcClient, metrics); err != nil {
			slog.Error("Send grpc metrics error", "err", err)
		}
		return
	}

	resp, err := SendBatchRequest(*client, config, metrics, pubKey)
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
			return
		}
		pubKey = pubKeyLoad
	}

	// gRPC-клиент создаётся при указании адреса через флаг -g или env GRPC.
	var grpcClient pb.MetricsClient
	if config.GrpcAddr != "" {
		conn, err := grpc.NewClient(
			config.GrpcAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			fmt.Printf("Grpc connect error: %v\n", err)
			return
		}
		defer func() { _ = conn.Close() }()
		grpcClient = pb.NewMetricsClient(conn)
		slog.Info("grpc client configured", "addr", config.GrpcAddr)
	}

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	defer tickerPoll.Stop()
	defer tickerReport.Stop()
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	var m runtime.MemStats
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	for {
		select {
		case t1 := <-tickerPoll.C:
			runtime.ReadMemStats(&m)
			slog.Info("Fast tic:", "time", t1)
		case t2 := <-tickerReport.C:
			sendData(
				ctx,
				client,
				grpcClient,
				config,
				&m,
				pubKey,
			)
			slog.Info("Slow tic:", "time", t2)
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			done := make(chan struct{})
			go func() {
				sendData(ctx, client, grpcClient, config, &m, pubKey)
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
