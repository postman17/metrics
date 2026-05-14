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

func TestUpdateHandlerSuccess(t *testing.T) {
	type want struct {
		code        int
		query       string
		metricType  string
		metricName  string
		metricValue string
		resultValue any
	}
	tests := []struct {
		name string
		data map[string]any
		want want
	}{
		{
			name: "gauge ok",
			data: make(map[string]any),
			want: want{
				code:        200,
				metricType:  "gauge",
				metricName:  "test",
				metricValue: "1.5",
				resultValue: 1.5,
			},
		},
		{
			name: "gauge ok exists",
			data: map[string]any{
				"test": 1.0,
			},
			want: want{
				code:        200,
				metricType:  "gauge",
				metricName:  "test",
				metricValue: "1.5",
				resultValue: 1.5,
			},
		},
		{
			name: "counter ok",
			data: make(map[string]any),
			want: want{
				code:        200,
				metricType:  "counter",
				metricName:  "test",
				metricValue: "1",
				resultValue: int64(1),
			},
		},
		{
			name: "counter ok exists",
			data: map[string]any{
				"test": 1,
			},
			want: want{
				code:        200,
				metricType:  "counter",
				metricName:  "test",
				metricValue: "1",
				resultValue: int64(1),
			},
		},
	}
	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodPost, "/update/gauge/test/1.5", nil)
		request.SetPathValue("type", tt.want.metricType)
		request.SetPathValue("name", tt.want.metricName)
		request.SetPathValue("value", tt.want.metricValue)
		w := httptest.NewRecorder()
		memory := repo.NewMemStorage()
		memory.Data = tt.data
		UpdateMetricPage(memory)(w, request)

		assert.Equal(t, memory.Data[tt.want.metricName], tt.want.resultValue)
		res := w.Result()
		assert.Equal(t, tt.want.code, res.StatusCode)

		defer res.Body.Close()
		_, err := io.ReadAll(res.Body)

		require.NoError(t, err)
	}
}

func TestUpdateHandlerErrors(t *testing.T) {
	type want struct {
		code        int
		query       string
		metricType  string
		metricName  string
		metricValue string
	}
	tests := []struct {
		name string
		data map[string]any
		want want
	}{
		{
			name: "without value",
			data: make(map[string]any),
			want: want{
				code:        404,
				metricType:  "gauge",
				metricName:  "test",
				metricValue: "",
			},
		},
		{
			name: "without name",
			data: make(map[string]any),
			want: want{
				code:        404,
				metricType:  "gauge",
				metricName:  "",
				metricValue: "1.5",
			},
		},
		{
			name: "without type",
			data: make(map[string]any),
			want: want{
				code:        404,
				metricType:  "",
				metricName:  "test",
				metricValue: "1.5",
			},
		},
		{
			name: "wrong type",
			data: make(map[string]any),
			want: want{
				code:        400,
				metricType:  "qwerty",
				metricName:  "test",
				metricValue: "1.5",
			},
		},
		{
			name: "wrong counter type exists",
			data: map[string]any{
				"test": int64(1),
			},
			want: want{
				code:        400,
				metricType:  "gauge",
				metricName:  "test",
				metricValue: "1.5",
			},
		},
		{
			name: "wrong gauge type exists",
			data: map[string]any{
				"test": 1.0,
			},
			want: want{
				code:        400,
				metricType:  "counter",
				metricName:  "test",
				metricValue: "1",
			},
		},
	}
	for _, tt := range tests {
		request := httptest.NewRequest(http.MethodPost, "/update/gauge/test/1.5", nil)
		request.SetPathValue("type", tt.want.metricType)
		request.SetPathValue("name", tt.want.metricName)
		request.SetPathValue("value", tt.want.metricValue)
		w := httptest.NewRecorder()
		memory := repo.NewMemStorage()
		memory.Data = tt.data
		UpdateMetricPage(memory)(w, request)

		res := w.Result()
		assert.Equal(t, tt.want.code, res.StatusCode)

		defer res.Body.Close()
		_, err := io.ReadAll(res.Body)

		require.NoError(t, err)
	}
}

func TestUpdateMetric(t *testing.T) {
	memory := repo.NewMemStorage()
	handler := http.HandlerFunc(UpdateMetric(memory))
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
			expectedBody: "",
		},
		{
			name:         "method_put",
			method:       http.MethodPut,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_delete",
			method:       http.MethodDelete,
			expectedCode: http.StatusMethodNotAllowed,
			expectedBody: "",
		},
		{
			name:         "method_post_without_body",
			method:       http.MethodPost,
			expectedCode: http.StatusInternalServerError,
			expectedBody: "",
		},
		{
			name:         "method_post_without_value",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "gauge"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			name:         "method_post_success",
			method:       http.MethodPost,
			body:         `{"id": "test", "type": "gauge", "value": 1.6}`,
			expectedCode: http.StatusOK,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader io.Reader
			if tc.body != "" {
				bodyReader = bytes.NewBufferString(tc.body)
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
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedCode, resp.StatusCode, "Response code didn't match expected")

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "error reading response body")

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(respBody))
			}
		})
	}
}
