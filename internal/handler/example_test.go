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

// ---------------------------------------------------------------------------
// POST /update/ — обновление одной метрики в JSON-формате
// ---------------------------------------------------------------------------

// ExampleUpdateMetric_gauge демонстрирует обновление gauge-метрики через
// JSON-эндпоинт POST /update/.
func ExampleUpdateMetric_gauge() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(UpdateMetric(storage)))
	defer srv.Close()

	// POST /update/  {"id":"Alloc","type":"gauge","value":42.5}
	body, _ := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: ptrFloat64(42.5),
	})

	resp, err := http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("status:", resp.StatusCode)
	fmt.Println("stored value:", storage.GetTypeValue("Alloc"))

	// Output:
	// status: 200
	// stored value: 42.5
}

// ExampleUpdateMetric_counter демонстрирует обновление counter-метрики через
// JSON-эндпоинт POST /update/. Счётчик накапливает значение.
func ExampleUpdateMetric_counter() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(UpdateMetric(storage)))
	defer srv.Close()

	// Первый запрос: counter = 5
	body1, _ := json.Marshal(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: ptrInt64(5),
	})
	resp1, _ := http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body1))
	resp1.Body.Close()

	// Второй запрос: counter += 3 → 8
	body2, _ := json.Marshal(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: ptrInt64(3),
	})
	resp2, _ := http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body2))
	resp2.Body.Close()

	fmt.Println("counter after two updates:", storage.GetTypeValue("PollCount"))

	// Output:
	// counter after two updates: 8
}

// ExampleUpdateMetric_badRequest демонстрирует ошибку при отправке gauge
// без обязательного поля value.
func ExampleUpdateMetric_badRequest() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(UpdateMetric(storage)))
	defer srv.Close()

	// value отсутствует — сервер вернёт 400
	body, _ := json.Marshal(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		// Value не задан
	})

	resp, _ := http.Post(srv.URL+"/update/", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	fmt.Println("status:", resp.StatusCode)

	// Output:
	// status: 400
}

// ---------------------------------------------------------------------------
// POST /value/ — получение метрики в JSON-формате
// ---------------------------------------------------------------------------

// ExampleGetMetricValue_gauge демонстрирует получение gauge-метрики через
// JSON-эндпоинт POST /value/.
func ExampleGetMetricValue_gauge() {
	storage := repo.NewMemStorage()
	storage.AddGauge("Alloc", 1024.56)

	srv := httptest.NewServer(http.HandlerFunc(GetMetricValue(storage)))
	defer srv.Close()

	reqBody, _ := json.Marshal(models.GetMetricRequest{ID: "Alloc", MType: models.Gauge})
	resp, err := http.Post(srv.URL+"/value/", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var m models.Metrics
	_ = json.Unmarshal(raw, &m)

	fmt.Println("status:", resp.StatusCode)
	fmt.Println("id:", m.ID, "type:", m.MType, "value:", *m.Value)

	// Output:
	// status: 200
	// id: Alloc type: gauge value: 1024.56
}

// ExampleGetMetricValue_counter демонстрирует получение counter-метрики через
// JSON-эндпоинт POST /value/.
func ExampleGetMetricValue_counter() {
	storage := repo.NewMemStorage()
	storage.AddCounter("PollCount", 42)

	srv := httptest.NewServer(http.HandlerFunc(GetMetricValue(storage)))
	defer srv.Close()

	reqBody, _ := json.Marshal(models.GetMetricRequest{ID: "PollCount", MType: models.Counter})
	resp, _ := http.Post(srv.URL+"/value/", "application/json", bytes.NewReader(reqBody))
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var m models.Metrics
	_ = json.Unmarshal(raw, &m)

	fmt.Println("id:", m.ID, "type:", m.MType, "delta:", *m.Delta)

	// Output:
	// id: PollCount type: counter delta: 42
}

// ExampleGetMetricValue_notFound демонстрирует ответ 404 при запросе
// несуществующей метрики.
func ExampleGetMetricValue_notFound() {
	storage := repo.NewMemStorage()
	srv := httptest.NewServer(http.HandlerFunc(GetMetricValue(storage)))
	defer srv.Close()

	reqBody, _ := json.Marshal(models.GetMetricRequest{ID: "Unknown", MType: models.Gauge})
	resp, _ := http.Post(srv.URL+"/value/", "application/json", bytes.NewReader(reqBody))
	defer resp.Body.Close()

	fmt.Println("status:", resp.StatusCode)

	// Output:
	// status: 404
}

// ---------------------------------------------------------------------------
// POST /updates/ — пакетное обновление метрик
// ---------------------------------------------------------------------------

// ExampleUpdatesMetric демонстрирует пакетное обновление нескольких метрик
// разных типов за один запрос POST /updates/.
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
	resp, err := http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer resp.Body.Close()

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

// ExampleUpdatesMetric_counterAccumulation демонстрирует накопление counter
// при повторных пакетных обновлениях.
func ExampleUpdatesMetric_counterAccumulation() {
	storage := repo.NewMemStorage()
	pub := &audit.Pub{}

	srv := httptest.NewServer(http.HandlerFunc(UpdatesMetric(storage, pub)))
	defer srv.Close()

	// Первый батч: PollCount = 5
	batch1 := models.MetricsList{
		{ID: "PollCount", MType: models.Counter, Delta: ptrInt64(5)},
	}
	body1, _ := json.Marshal(batch1)
	resp1, _ := http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body1))
	resp1.Body.Close()

	// Второй батч: PollCount += 7 → 12
	batch2 := models.MetricsList{
		{ID: "PollCount", MType: models.Counter, Delta: ptrInt64(7)},
	}
	body2, _ := json.Marshal(batch2)
	resp2, _ := http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body2))
	resp2.Body.Close()

	fmt.Println("PollCount after two batches:", storage.GetTypeValue("PollCount"))

	// Output:
	// PollCount after two batches: 12
}

// ExampleUpdatesMetric_invalidMetric демонстрирует ошибку 400 при отправке
// пакета с невалидной метрикой (пустой ID).
func ExampleUpdatesMetric_invalidMetric() {
	storage := repo.NewMemStorage()
	pub := &audit.Pub{}

	srv := httptest.NewServer(http.HandlerFunc(UpdatesMetric(storage, pub)))
	defer srv.Close()

	// ID пустой — сервер вернёт 400, вся транзакция отменена
	batch := models.MetricsList{
		{ID: "", MType: models.Gauge, Value: ptrFloat64(1.0)},
	}

	body, _ := json.Marshal(batch)
	resp, _ := http.Post(srv.URL+"/updates/", "application/json", bytes.NewReader(body))
	defer resp.Body.Close()

	fmt.Println("status:", resp.StatusCode)

	// Output:
	// status: 400
}

// ---------------------------------------------------------------------------
// вспомогательные функции
// ---------------------------------------------------------------------------

func ptrFloat64(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64       { return &v }
