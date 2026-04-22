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
