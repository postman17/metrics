package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	audit "github.com/postman17/metrics/internal/audit"
	models "github.com/postman17/metrics/internal/model"
	repo "github.com/postman17/metrics/internal/repository"
)

func checkPostResp(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("nil response from http.Post")
	}
	return resp, nil
}

func ExampleUpdateMetric_gauge() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(UpdateMetric(storage)))
	defer srv.Close()

	body, _ := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: ptrFloat64(42.5),
	})

	resp, err := checkPostResp(http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println("status:", resp.StatusCode)
	fmt.Println("stored value:", storage.GetTypeValue("Alloc"))

	// Output:
	// status: 200
	// stored value: 42.5
}

func ExampleUpdateMetric_counter() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(UpdateMetric(storage)))
	defer srv.Close()

	body1, _ := json.Marshal(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: ptrInt64(5),
	})
	resp1, err := checkPostResp(http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body1)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = resp1.Body.Close()

	body2, _ := json.Marshal(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: ptrInt64(3),
	})
	resp2, err := checkPostResp(http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body2)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = resp2.Body.Close()

	fmt.Println("counter after two updates:", storage.GetTypeValue("PollCount"))

	// Output:
	// counter after two updates: 8
}

func ExampleUpdateMetric_badRequest() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(UpdateMetric(storage)))
	defer srv.Close()

	body, _ := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
	})

	resp, err := checkPostResp(http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println("status:", resp.StatusCode)

	// Output:
	// status: 400
}

func ExampleGetMetricValue_gauge() {
	storage := repo.NewMemStorage()
	storage.AddGauge("Alloc", 1024.56)

	srv := httptest.NewServer(http.HandlerFunc(GetMetricValue(storage)))
	defer srv.Close()

	reqBody, _ := json.Marshal(models.GetMetricRequest{ID: "Alloc", MType: models.Gauge})
	resp, err := checkPostResp(http.Post(srv.URL+"/value/", "application/json", bytes.NewReader(reqBody)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(resp.Body)

	var m models.Metrics
	_ = json.Unmarshal(raw, &m)

	fmt.Println("status:", resp.StatusCode)
	fmt.Println("id:", m.ID, "type:", m.MType, "value:", *m.Value)

	// Output:
	// status: 200
	// id: Alloc type: gauge value: 1024.56
}

func ExampleGetMetricValue_counter() {
	storage := repo.NewMemStorage()
	storage.AddCounter("PollCount", 42)

	srv := httptest.NewServer(http.HandlerFunc(GetMetricValue(storage)))
	defer srv.Close()

	reqBody, _ := json.Marshal(models.GetMetricRequest{ID: "PollCount", MType: models.Counter})
	resp, err := checkPostResp(http.Post(srv.URL+"/value/", "application/json", bytes.NewReader(reqBody)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	raw, _ := io.ReadAll(resp.Body)

	var m models.Metrics
	_ = json.Unmarshal(raw, &m)

	fmt.Println("id:", m.ID, "type:", m.MType, "delta:", *m.Delta)

	// Output:
	// id: PollCount type: counter delta: 42
}

func ExampleGetMetricValue_notFound() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(GetMetricValue(storage)))
	defer srv.Close()

	reqBody, _ := json.Marshal(models.GetMetricRequest{ID: "Unknown", MType: models.Gauge})
	resp, err := checkPostResp(http.Post(srv.URL+"/value/", "application/json", bytes.NewReader(reqBody)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println("status:", resp.StatusCode)

	// Output:
	// status: 404
}

func ExampleUpdatesMetric() {
	storage := repo.NewMemStorage()
	pub := &audit.Pub{}

	srv := httptest.NewServer(http.HandlerFunc(UpdatesMetric(storage, pub)))
	defer srv.Close()

	batch := models.MetricsList{
		{ID: "Alloc", MType: models.Gauge, Value: ptrFloat64(2048.0)},
		{ID: "HeapSys", MType: models.Gauge, Value: ptrFloat64(4096.5)},
		{ID: "PollCount", MType: models.Counter, Delta: ptrInt64(1)},
	}

	body, _ := json.Marshal(batch)
	resp, err := checkPostResp(http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println("status:", resp.StatusCode)
	fmt.Println("Alloc:", storage.GetTypeValue("Alloc"))
	fmt.Println("HeapSys:", storage.GetTypeValue("HeapSys"))
	fmt.Println("PollCount:", storage.GetTypeValue("PollCount"))

	// Output:
	// status: 200
	// Alloc: 2048
	// HeapSys: 4096.5
	// PollCount: 1
}

func ExampleUpdatesMetric_counterAccumulation() {
	storage := repo.NewMemStorage()
	pub := &audit.Pub{}

	srv := httptest.NewServer(http.HandlerFunc(UpdatesMetric(storage, pub)))
	defer srv.Close()

	batch1 := models.MetricsList{
		{ID: "PollCount", MType: models.Counter, Delta: ptrInt64(5)},
	}
	body1, _ := json.Marshal(batch1)
	resp1, err := checkPostResp(http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body1)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = resp1.Body.Close()

	batch2 := models.MetricsList{
		{ID: "PollCount", MType: models.Counter, Delta: ptrInt64(7)},
	}
	body2, _ := json.Marshal(batch2)
	resp2, err := checkPostResp(http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body2)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	_ = resp2.Body.Close()

	fmt.Println("PollCount after two batches:", storage.GetTypeValue("PollCount"))

	// Output:
	// PollCount after two batches: 12
}

func ExampleUpdatesMetric_invalidMetric() {
	storage := repo.NewMemStorage()
	pub := &audit.Pub{}

	srv := httptest.NewServer(http.HandlerFunc(UpdatesMetric(storage, pub)))
	defer srv.Close()

	batch := models.MetricsList{
		{ID: "", MType: models.Gauge, Value: ptrFloat64(1.0)},
	}

	body, _ := json.Marshal(batch)
	resp, err := checkPostResp(http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body)))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Println("status:", resp.StatusCode)

	// Output:
	// status: 400
}

func ptrFloat64(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64       { return &v }
