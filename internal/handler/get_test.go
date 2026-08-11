package handler

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/require"
	repo "github.com/postman17/metrics/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestGetValueHandlerSuccess(t *testing.T) {
	type want struct {
		code        int
		query       string
		metricType  string
		metricName  string
		resultValue string
	}
	tests := []struct {
		name string
		data map[string]any
		want want
	}{
		{
			name: "counter ok",
			data: map[string]any{
				"test": 1,
			},
			want: want{
				code:        200,
				metricType:  "counter",
				metricName:  "test",
				resultValue: "1",
			},
		},
		{
			name: "gauge ok",
			data: map[string]any{
				"test": 1.5,
			},
			want: want{
				code:        200,
				metricType:  "gauge",
				metricName:  "test",
				resultValue: "1.5",
			},
		},
	}
	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
		request.SetPathValue("type", tt.want.metricType)
		request.SetPathValue("name", tt.want.metricName)
		w := httptest.NewRecorder()
		memory := repo.NewMemStorageWithData(tt.data)
		GetMetricValuePage(memory)(w, request)
		res := w.Result()
		body, _ := io.ReadAll(res.Body)
		bodyStr := string(body)
		assert.Equal(t, bodyStr, tt.want.resultValue)
		assert.Equal(t, tt.want.code, res.StatusCode)

		defer func() { _ = res.Body.Close() }()
		_, err := io.ReadAll(res.Body)

		require.NoError(t, err)
	}
}

func TestGetValueHandlerErrors(t *testing.T) {
	type want struct {
		code       int
		query      string
		metricType string
		metricName string
	}
	tests := []struct {
		name string
		data map[string]any
		want want
	}{
		{
			name: "without value in memory",
			data: make(map[string]any),
			want: want{
				code:       404,
				metricType: "gauge",
				metricName: "test",
			},
		},
		{
			name: "without name",
			data: make(map[string]any),
			want: want{
				code:       404,
				metricType: "gauge",
				metricName: "",
			},
		},
		{
			name: "without type",
			data: make(map[string]any),
			want: want{
				code:       404,
				metricType: "",
				metricName: "test",
			},
		},
	}
	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodGet, "/update/gauge/test/1.5", nil)
		request.SetPathValue("type", tt.want.metricType)
		request.SetPathValue("name", tt.want.metricName)

		w := httptest.NewRecorder()
		memory := repo.NewMemStorageWithData(tt.data)
		GetMetricValuePage(memory)(w, request)

		res := w.Result()
		assert.Equal(t, tt.want.code, res.StatusCode)

		defer func() { _ = res.Body.Close() }()
		_, err := io.ReadAll(res.Body)

		require.NoError(t, err)
	}
}

func TestGetMetric(t *testing.T) {
	memory := repo.NewMemStorage()
	handler := http.HandlerFunc(GetMetricValue(memory))
	srv := httptest.NewServer(handler)
	defer srv.Close()
	type mem struct {
		metricName  string
		metricValue float64
	}
	testCases := []struct {
		name         string
		method       string
		body         string
		expectedCode int
		expectedBody string
		hasMem       bool
		mem          mem
	}{
		{
			name:         "method_get",
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
			hasMem:       false,
			mem:          mem{},
		},
		{
			name:         "method_put",
			method:       http.MethodPut,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
			hasMem:       false,
			mem:          mem{},
		},
		{
			name:         "method_delete",
			method:       http.MethodDelete,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
			hasMem:       false,
			mem:          mem{},
		},
		{
			name:         "method_post_without_body",
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			hasMem:       false,
		},
		{
			name:         "method_post_without_value",
			method:       http.MethodPost,
			body:         `{"id": "test"}`,
			expectedCode: http.StatusNotFound,
			expectedBody: "",
			hasMem:       false,
			mem:          mem{},
		},
		{
			name:         "method_post_not_type",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "test"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			hasMem:       false,
			mem:          mem{},
		},
		{
			name:         "method_post_success",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "gauge"}`,
			expectedCode: http.StatusOK,
			expectedBody: `{"id": "test", "type": "gauge", "value": 1.6}`,
			hasMem:       true,
			mem: mem{
				metricName:  "test",
				metricValue: 1.6,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = bytes.NewBufferString(tc.body)
			}

			if tc.hasMem {
				memory.AddGauge(tc.mem.metricName, tc.mem.metricValue)
			}
			req, err := http.NewRequest(tc.method, srv.URL+"/update", bodyReader)
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
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tc.expectedCode, resp.StatusCode, "Response code didn't match expected")

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "error reading response body")

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(respBody))
			}
		})
	}
}
