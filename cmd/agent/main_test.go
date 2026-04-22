package main

import (
	"net/http"
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/require"
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

	resp, err := SendGaugeData(*client, "test", 1.5)

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

	resp, err := SendCounterData(*client, "test", 1)

	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, resp.StatusCode, http.StatusOK)
}
