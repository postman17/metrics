package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"
)

func SendGaugeData(client http.Client, name string, data float64) (*http.Response, error) {
	url := fmt.Sprintf("%s/update/gauge/%s/%v", flagRunAddr, name, data)
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer resp.Body.Close()
	return resp, err
}

func SendCounterData(client http.Client, name string, data int64) (*http.Response, error) {
	url := fmt.Sprintf("%s/update/counter/%s/%v", flagRunAddr, name, data)
	resp, err := client.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer resp.Body.Close()
	return resp, err
}

func main() {
	parseFlags()

	tickerPoll := time.NewTicker(time.Duration(flagPollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(flagReportInterval) * time.Second)
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
			SendGaugeData(*client, "Alloc", float64(m.Alloc))
			SendGaugeData(*client, "BuckHashSys", float64(m.BuckHashSys))
			SendGaugeData(*client, "Frees", float64(m.Frees))
			SendGaugeData(*client, "GCCPUFraction", float64(m.GCCPUFraction))
			SendGaugeData(*client, "GCSys", float64(m.GCSys))
			SendGaugeData(*client, "HeapAlloc", float64(m.HeapAlloc))
			SendGaugeData(*client, "HeapIdle", float64(m.HeapIdle))
			SendGaugeData(*client, "HeapInuse", float64(m.HeapInuse))
			SendGaugeData(*client, "HeapObjects", float64(m.HeapObjects))
			SendGaugeData(*client, "HeapReleased", float64(m.HeapReleased))
			SendGaugeData(*client, "HeapSys", float64(m.HeapSys))
			SendGaugeData(*client, "LastGC", float64(m.LastGC))
			SendGaugeData(*client, "Lookups", float64(m.Lookups))
			SendGaugeData(*client, "MCacheInuse", float64(m.MCacheInuse))
			SendGaugeData(*client, "MCacheSys", float64(m.MCacheSys))
			SendGaugeData(*client, "MSpanInuse", float64(m.MSpanInuse))
			SendGaugeData(*client, "MSpanSys", float64(m.MSpanSys))
			SendGaugeData(*client, "Mallocs", float64(m.Mallocs))
			SendGaugeData(*client, "NextGC", float64(m.NextGC))
			SendGaugeData(*client, "NumForcedGC", float64(m.NumForcedGC))
			SendGaugeData(*client, "NumGC", float64(m.NumGC))
			SendGaugeData(*client, "OtherSys", float64(m.OtherSys))
			SendGaugeData(*client, "PauseTotalNs", float64(m.PauseTotalNs))
			SendGaugeData(*client, "StackInuse", float64(m.StackInuse))
			SendGaugeData(*client, "StackSys", float64(m.StackSys))
			SendGaugeData(*client, "Sys", float64(m.Sys))
			SendGaugeData(*client, "TotalAlloc", float64(m.TotalAlloc))
			SendGaugeData(*client, "RandomValue", rand.Float64())
			SendCounterData(*client, "PollCount", 1)
			fmt.Println("Slow tic:", t2)
		}
	}
}
