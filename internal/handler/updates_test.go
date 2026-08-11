package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mailru/easyjson"
	"github.com/postman17/metrics/internal/audit"
	models "github.com/postman17/metrics/internal/model"
	repo "github.com/postman17/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdatesMetric(t *testing.T) {
	memory := repo.NewMemStorage()
	pub := &audit.Pub{}
	handler := http.HandlerFunc(UpdatesMetric(memory, pub))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	testCases := []struct {
		name         string
		method       string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "method_get",
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "method_put",
			method:       http.MethodPut,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "method_delete",
			method:       http.MethodDelete,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "method_post_without_body",
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "method_post_invalid_json",
			method:       http.MethodPost,
			body:         `{invalid`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "method_post_empty_id",
			method:       http.MethodPost,
			body:         `[{"id": "", "type": "gauge", "value": 1.0}]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "method_post_invalid_type",
			method:       http.MethodPost,
			body:         `[{"id": "test", "type": "unknown", "value": 1.0}]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "method_post_gauge_without_value",
			method:       http.MethodPost,
			body:         `[{"id": "test", "type": "gauge"}]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "method_post_counter_without_delta",
			method:       http.MethodPost,
			body:         `[{"id": "test", "type": "counter"}]`,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "method_post_empty_list",
			method:       http.MethodPost,
			body:         `[]`,
			expectedCode: http.StatusOK,
			expectedBody: `{}`,
		},
		{
			name:         "method_post_success",
			method:       http.MethodPost,
			body:         `[{"id": "Alloc", "type": "gauge", "value": 1.6}, {"id": "PollCount", "type": "counter", "delta": 5}]`,
			expectedCode: http.StatusOK,
			expectedBody: `{}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = bytes.NewBufferString(tc.body)
			}

			req, err := http.NewRequest(tc.method, srv.URL+"/updates", bodyReader)
			if err != nil {
				t.Fatalf("Не удалось создать запрос: %v", err)
			}

			if tc.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := srv.Client().Do(req)
			if !assert.NoError(t, err, "error making HTTP request") {
				return
			}
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(respBody))
			}
		})
	}
}

func TestUpdatesMetricSuccess(t *testing.T) {
	memory := repo.NewMemStorage()
	pub := &audit.Pub{}

	body := `[{"id": "Alloc", "type": "gauge", "value": 1.6}, {"id": "PollCount", "type": "counter", "delta": 5}]`
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdatesMetric(memory, pub)(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	respBody, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(respBody))

	assert.Equal(t, 1.6, memory.GetTypeValue("Alloc"))
	assert.Equal(t, int64(5), memory.GetTypeValue("PollCount"))
}

func TestUpdatesMetricCounterAccumulation(t *testing.T) {
	memory := repo.NewMemStorageWithData(map[string]any{
		"PollCount": int64(10),
	})
	pub := &audit.Pub{}

	body := `[{"id": "PollCount", "type": "counter", "delta": 3}]`
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdatesMetric(memory, pub)(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, int64(13), memory.GetTypeValue("PollCount"))
}

func BenchmarkUpdatesMetric(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"1_metric", 1},
		{"10_metrics", 10},
		{"100_metrics", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			list := make(models.MetricsList, 0, bm.count)
			for i := 0; i < bm.count; i++ {
				v := float64(i) + 0.5
				list = append(list, models.Metrics{
					ID:    fmt.Sprintf("gauge_%d", i),
					MType: models.Gauge,
					Value: &v,
				})
			}

			body, err := easyjson.Marshal(list)
			if err != nil {
				b.Fatalf("failed to marshal: %v", err)
			}

			memory := repo.NewMemStorage()
			pub := &audit.Pub{}
			handler := UpdatesMetric(memory, pub)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				handler(w, req)
				if w.Code != http.StatusOK {
					b.Fatalf("unexpected status code: %d", w.Code)
				}
			}
		})
	}
}

func BenchmarkUpdatesMetricMixed(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"10_mixed", 10},
		{"100_mixed", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			list := make(models.MetricsList, 0, bm.count)
			for i := 0; i < bm.count; i++ {
				if i%2 == 0 {
					v := float64(i) + 0.5
					list = append(list, models.Metrics{
						ID:    fmt.Sprintf("gauge_%d", i),
						MType: models.Gauge,
						Value: &v,
					})
				} else {
					d := int64(i)
					list = append(list, models.Metrics{
						ID:    fmt.Sprintf("counter_%d", i),
						MType: models.Counter,
						Delta: &d,
					})
				}
			}

			body, err := easyjson.Marshal(list)
			if err != nil {
				b.Fatalf("failed to marshal: %v", err)
			}

			memory := repo.NewMemStorage()
			pub := &audit.Pub{}
			handler := UpdatesMetric(memory, pub)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				handler(w, req)
				if w.Code != http.StatusOK {
					b.Fatalf("unexpected status code: %d", w.Code)
				}
			}
		})
	}
}

func BenchmarkUpdatesMetricCounterAccumulation(b *testing.B) {
	d := int64(1)
	list := models.MetricsList{
		{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &d,
		},
	}

	body, err := easyjson.Marshal(list)
	if err != nil {
		b.Fatalf("failed to marshal: %v", err)
	}

	memory := repo.NewMemStorage()
	pub := &audit.Pub{}
	handler := UpdatesMetric(memory, pub)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

func BenchmarkUpdatesMetricMethodNotAllowed(b *testing.B) {
	memory := repo.NewMemStorage()
	pub := &audit.Pub{}
	handler := UpdatesMetric(memory, pub)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/updates", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}

func BenchmarkUpdatesMetricInvalidJSON(b *testing.B) {
	body := []byte(`{invalid`)

	memory := repo.NewMemStorage()
	pub := &audit.Pub{}
	handler := UpdatesMetric(memory, pub)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusBadRequest {
			b.Fatalf("unexpected status code: %d", w.Code)
		}
	}
}
