package handler

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	repo "github.com/postman17/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestUpdatesMetric(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	memory := repo.NewMemStorage(ctx, true, "", false, &sql.DB{}, false)
	handler := http.HandlerFunc(UpdatesMetric(memory))
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	memory := repo.NewMemStorage(ctx, true, "", false, &sql.DB{}, false)

	body := `[{"id": "Alloc", "type": "gauge", "value": 1.6}, {"id": "PollCount", "type": "counter", "delta": 5}]`
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdatesMetric(memory)(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	respBody, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.JSONEq(t, `{}`, string(respBody))

	assert.Equal(t, 1.6, memory.Data["Alloc"])
	assert.Equal(t, int64(5), memory.Data["PollCount"])
}

func TestUpdatesMetricCounterAccumulation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	memory := repo.NewMemStorage(ctx, true, "", false, &sql.DB{}, false)
	memory.Data["PollCount"] = int64(10)

	body := `[{"id": "PollCount", "type": "counter", "delta": 3}]`
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	UpdatesMetric(memory)(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, int64(13), memory.Data["PollCount"])
}
