package handler

import (
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
		memory := repo.NewMemStorage()
		memory.Data = tt.data
		GetMetricValuePage(memory)(w, request)
		res := w.Result()
		body, _ := io.ReadAll(res.Body)
		bodyStr := string(body)
		assert.Equal(t, bodyStr, tt.want.resultValue)
		assert.Equal(t, tt.want.code, res.StatusCode)

		defer res.Body.Close()
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
		memory := repo.NewMemStorage()
		memory.Data = tt.data
		GetMetricValuePage(memory)(w, request)

		res := w.Result()
		assert.Equal(t, tt.want.code, res.StatusCode)

		defer res.Body.Close()
		_, err := io.ReadAll(res.Body)

		require.NoError(t, err)
	}
}
