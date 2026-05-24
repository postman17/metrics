package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
)

func SendRequest(client http.Client, metric models.Metrics, url string) (*http.Response, error) {
	jsonData, err := easyjson.Marshal(metric)
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return nil, err
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err = gzWriter.Write(jsonData)
	if err != nil {
		slog.Error("Gzip error: %v\n", "err", err)
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		slog.Error("Error gzip compress: %v\n", "err", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		slog.Error("Create request error: %v\n", "err", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Send request error: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}

func SendGaugeData(client http.Client, config Config, name string, data float64) (*http.Response, error) {
	url := fmt.Sprintf("%s/update/", config.RunAddr)
	metric := models.Metrics{
		ID:    name,
		MType: "gauge",
		Value: &data,
	}

	return SendRequest(client, metric, url)
}

func SendCounterData(client http.Client, config Config, name string, data int64) (*http.Response, error) {
	url := fmt.Sprintf("%s/update/", config.RunAddr)
	metric := models.Metrics{
		ID:    name,
		MType: "counter",
		Delta: &data,
	}

	return SendRequest(client, metric, url)
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
			SendGaugeData(*client, config, "Alloc", float64(m.Alloc))
			SendGaugeData(*client, config, "BuckHashSys", float64(m.BuckHashSys))
			SendGaugeData(*client, config, "Frees", float64(m.Frees))
			SendGaugeData(*client, config, "GCCPUFraction", float64(m.GCCPUFraction))
			SendGaugeData(*client, config, "GCSys", float64(m.GCSys))
			SendGaugeData(*client, config, "HeapAlloc", float64(m.HeapAlloc))
			SendGaugeData(*client, config, "HeapIdle", float64(m.HeapIdle))
			SendGaugeData(*client, config, "HeapInuse", float64(m.HeapInuse))
			SendGaugeData(*client, config, "HeapObjects", float64(m.HeapObjects))
			SendGaugeData(*client, config, "HeapReleased", float64(m.HeapReleased))
			SendGaugeData(*client, config, "HeapSys", float64(m.HeapSys))
			SendGaugeData(*client, config, "LastGC", float64(m.LastGC))
			SendGaugeData(*client, config, "Lookups", float64(m.Lookups))
			SendGaugeData(*client, config, "MCacheInuse", float64(m.MCacheInuse))
			SendGaugeData(*client, config, "MCacheSys", float64(m.MCacheSys))
			SendGaugeData(*client, config, "MSpanInuse", float64(m.MSpanInuse))
			SendGaugeData(*client, config, "MSpanSys", float64(m.MSpanSys))
			SendGaugeData(*client, config, "Mallocs", float64(m.Mallocs))
			SendGaugeData(*client, config, "NextGC", float64(m.NextGC))
			SendGaugeData(*client, config, "NumForcedGC", float64(m.NumForcedGC))
			SendGaugeData(*client, config, "NumGC", float64(m.NumGC))
			SendGaugeData(*client, config, "OtherSys", float64(m.OtherSys))
			SendGaugeData(*client, config, "PauseTotalNs", float64(m.PauseTotalNs))
			SendGaugeData(*client, config, "StackInuse", float64(m.StackInuse))
			SendGaugeData(*client, config, "StackSys", float64(m.StackSys))
			SendGaugeData(*client, config, "Sys", float64(m.Sys))
			SendGaugeData(*client, config, "TotalAlloc", float64(m.TotalAlloc))
			SendGaugeData(*client, config, "RandomValue", rand.Float64())
			SendCounterData(*client, config, "PollCount", 1)
			slog.Info("Slow tic:", "time", t2)
		}
	}
}
