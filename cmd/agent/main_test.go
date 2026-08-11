package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockRoundTripFunc func(req *http.Request) *http.Response

func (f MockRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestSendBatchRequest_MockTransport(t *testing.T) {
	client := &http.Client{
		Transport: MockRoundTripFunc(func(req *http.Request) *http.Response {
			assert.Equal(t, http.MethodPost, req.Method)
			assert.True(t, strings.HasSuffix(req.URL.Path, "/updates/"))
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, "gzip", req.Header.Get("Content-Encoding"))
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}
		}),
	}

	config := Config{RunAddr: "http://localhost:8080"}
	metrics := models.MetricsList{
		gaugeMetric("test_gauge", 1.5),
		counterMetric("test_counter", 1),
	}

	resp, err := SendBatchRequest(*client, config, metrics)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSendBatchRequest_GzipPayload(t *testing.T) {
	var received models.MetricsList

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/updates/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		gzReader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer func() { _ = gzReader.Close() }()

		body, err := io.ReadAll(gzReader)
		require.NoError(t, err)
		require.NoError(t, easyjson.Unmarshal(body, &received))

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := Config{RunAddr: ts.URL}
	metrics := models.MetricsList{
		gaugeMetric("test_gauge", 123.45),
		counterMetric("test_counter", 42),
	}

	resp, err := SendBatchRequest(http.Client{}, config, metrics)
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, received, 2)

	byID := make(map[string]models.Metrics, len(received))
	for _, m := range received {
		byID[m.ID] = m
	}

	gauge, ok := byID["test_gauge"]
	require.True(t, ok)
	assert.Equal(t, "gauge", gauge.MType)
	require.NotNil(t, gauge.Value)
	assert.Equal(t, 123.45, *gauge.Value)

	counter, ok := byID["test_counter"]
	require.True(t, ok)
	assert.Equal(t, "counter", counter.MType)
	require.NotNil(t, counter.Delta)
	assert.Equal(t, int64(42), *counter.Delta)
}

func TestRuntimeMetrics(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := runtimeMetrics(&m)
	require.Len(t, metrics, 29)

	byID := make(map[string]models.Metrics, len(metrics))
	for _, metric := range metrics {
		byID[metric.ID] = metric
	}

	assert.Equal(t, "gauge", byID["Alloc"].MType)
	require.NotNil(t, byID["Alloc"].Value)

	assert.Equal(t, "counter", byID["PollCount"].MType)
	require.NotNil(t, byID["PollCount"].Delta)
	assert.Equal(t, int64(1), *byID["PollCount"].Delta)

	_, ok := byID["RandomValue"]
	assert.True(t, ok)
}
