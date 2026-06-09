package sha256

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware_ValidHash(t *testing.T) {
	key := "secret"
	body := []byte(`{"id":"test","type":"gauge","value":1}`)

	req := httptest.NewRequest(http.MethodPost, "/update/", io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("HashSHA256", getHash(key))

	rr := httptest.NewRecorder()
	handler := Middleware(key)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, getHash(key), rr.Header().Get("HashSHA256"))
}

func TestMiddleware_EmptyKeySkipsCheck(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update/", nil)

	rr := httptest.NewRecorder()
	called := false
	handler := Middleware("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddleware_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update/", io.NopCloser(bytes.NewReader([]byte(`{}`))))

	rr := httptest.NewRecorder()
	called := false
	handler := Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, getHash("secret"), rr.Header().Get("HashSHA256"))
}

func TestMiddleware_InvalidHash(t *testing.T) {
	body := []byte(`{"id":"test","type":"gauge","value":1}`)
	req := httptest.NewRequest(http.MethodPost, "/update/", io.NopCloser(bytes.NewReader(body)))
	req.Header.Set("HashSHA256", "invalid")

	rr := httptest.NewRecorder()
	handler := Middleware("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetHash(t *testing.T) {
	hash := getHash("key")
	require.NotEmpty(t, hash)
	assert.Equal(t, hash, getHash("key"))
	assert.NotEqual(t, hash, getHash("other"))
}
