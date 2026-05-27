package main

import (
	"net/http"
	"testing"

	"compress/gzip"
	"encoding/json"
	"io"
	"net/http/httptest"

	"github.com/go-jose/go-jose/v4/testutils/require"
	models "github.com/postman17/metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

type MockRoundTripFunc func(req *http.Request) *http.Response

func (f MockRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestGaugeClient(t *testing.T) {
	client := &http.Client{
		Transport: MockRoundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
				Header:     make(http.Header),
			}
		}),
	}
	config := Config{}
	resp, err := SendGaugeData(*client, config, "test", 1.5)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestCounterClient(t *testing.T) {
	client := &http.Client{
		Transport: MockRoundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
				Header:     make(http.Header),
			}
		}),
	}
	config := Config{}
	resp, err := SendCounterData(*client, config, "test", 1)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusOK)
}

func TestGzipSendMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		gzReader, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gzReader.Close()

		body, err := io.ReadAll(gzReader)
		require.NoError(t, err)

		var receivedMetric models.Metrics
		err = json.Unmarshal(body, &receivedMetric)
		require.NoError(t, err)

		if receivedMetric.MType == "gauge" {
			assert.Equal(t, "test_gauge", receivedMetric.ID)
			assert.Equal(t, 123.45, *receivedMetric.Value)
		} else if receivedMetric.MType == "counter" {
			assert.Equal(t, "test_counter", receivedMetric.ID)
			assert.Equal(t, int64(42), *receivedMetric.Delta)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := Config{RunAddr: ts.URL}
	client := http.Client{}

	t.Run("Test SendGaugeData", func(t *testing.T) {
		resp, err := SendGaugeData(client, config, "test_gauge", 123.45)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Test SendCounterData", func(t *testing.T) {
		resp, err := SendCounterData(client, config, "test_counter", 42)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
