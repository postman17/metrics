package audit

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestURLSubscriber_getID(t *testing.T) {
	s := &URLSubscriber{ID: "url-1", URL: "http://example.com"}
	assert.Equal(t, "url-1", s.getID())
}

func TestURLSubscriber_send(t *testing.T) {
	t.Run("sends post successfully", func(t *testing.T) {
		var receivedBody string
		var receivedContentType string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedContentType = r.Header.Get("Content-Type")
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		s := &URLSubscriber{ID: "url-1", URL: srv.URL}
		err := s.send("192.168.1.1", []string{"cpu", "mem"})
		require.NoError(t, err)

		assert.Equal(t, "application/json", receivedContentType)
		assert.Contains(t, receivedBody, `"ip_address":"192.168.1.1"`)
		assert.Contains(t, receivedBody, `"metrics":["cpu","mem"]`)
		assert.Contains(t, receivedBody, `"ts":`)
	})

	t.Run("returns error on bad status code", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		s := &URLSubscriber{ID: "url-2", URL: srv.URL}
		err := s.send("10.0.0.1", []string{"cpu"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bad status code")
	})

	t.Run("returns error on unreachable url", func(t *testing.T) {
		s := &URLSubscriber{ID: "url-3", URL: "http://127.0.0.1:1"}
		err := s.send("10.0.0.1", []string{"cpu"})
		assert.Error(t, err)
	})

	t.Run("returns error on invalid url", func(t *testing.T) {
		s := &URLSubscriber{ID: "url-4", URL: "http://invalid-host-that-does-not-exist.local/audit"}
		err := s.send("10.0.0.1", []string{"cpu"})
		assert.Error(t, err)
	})

	t.Run("status 2xx accepted", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
		}{
			{name: "200 OK", statusCode: http.StatusOK},
			{name: "201 Created", statusCode: http.StatusCreated},
			{name: "204 No Content", statusCode: http.StatusNoContent},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
				}))
				defer srv.Close()

				s := &URLSubscriber{ID: "url-5", URL: srv.URL}
				err := s.send("10.0.0.1", []string{"cpu"})
				assert.NoError(t, err)
			})
		}
	})

	t.Run("status 3xx and 4xx rejected", func(t *testing.T) {
		tests := []struct {
			name       string
			statusCode int
		}{
			{name: "301 Moved", statusCode: http.StatusMovedPermanently},
			{name: "400 Bad Request", statusCode: http.StatusBadRequest},
			{name: "403 Forbidden", statusCode: http.StatusForbidden},
			{name: "404 Not Found", statusCode: http.StatusNotFound},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.statusCode)
				}))
				defer srv.Close()

				s := &URLSubscriber{ID: "url-6", URL: srv.URL}
				err := s.send("10.0.0.1", []string{"cpu"})
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), "bad status code"))
			})
		}
	})
}
