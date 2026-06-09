package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	repo "github.com/postman17/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestGetMainPage(t *testing.T) {
	storage := repo.NewMemStorageWithData(map[string]any{
		"GaugeMetric":   1.23,
		"CounterMetric": int64(10),
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler := GetMainPage(storage)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, rr.Code, http.StatusOK)

	contentType := rr.Header().Get("Content-Type")
	assert.Equal(t, contentType, "text/html")

	// Проверяем, что в теле ответа есть наши метрики
	body := rr.Body.String()
	expectedSubstrings := []string{
		"<h1>Current Metrics</h1>",
		"<li><strong>GaugeMetric</strong>: 1.23</li>",
		"<li><strong>CounterMetric</strong>: 10</li>",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(body, s) {
			t.Errorf("expected body to contain %q", s)
		}
	}
}
