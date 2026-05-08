package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"
)

func SendGaugeData(client http.Client, config Config, name string, data float64) (*http.Response, error) {
	url := fmt.Sprintf("%s/update/gauge/%s/%v", config.RunAddr, name, data)
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer resp.Body.Close()
	return resp, err
}

func SendCounterData(client http.Client, config Config, name string, data int64) (*http.Response, error) {
	url := fmt.Sprintf("%s/update/counter/%s/%v", config.RunAddr, name, data)
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer resp.Body.Close()
	return resp, err
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
			fmt.Println("Fast tic:", t1)
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
			fmt.Println("Slow tic:", t2)
		}
	}
}
