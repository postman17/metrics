package audit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockObserver struct {
	id          string
	lastIP      string
	lastMetrics []string
	sendErr     error
}

func (m *mockObserver) send(ip string, metrics []string) error {
	m.lastIP = ip
	m.lastMetrics = metrics
	return m.sendErr
}

func (m *mockObserver) getID() string {
	return m.id
}

func TestPub_Register(t *testing.T) {
	tests := []struct {
		name      string
		pub       *Pub
		observers []observer
		wantLen   int
	}{
		{
			name:      "register single observer on nil map",
			pub:       &Pub{},
			observers: []observer{&mockObserver{id: "obs1"}},
			wantLen:   1,
		},
		{
			name:      "register multiple observers",
			pub:       &Pub{},
			observers: []observer{&mockObserver{id: "obs1"}, &mockObserver{id: "obs2"}},
			wantLen:   2,
		},
		{
			name:      "register overwrites same id",
			pub:       &Pub{},
			observers: []observer{&mockObserver{id: "obs1"}, &mockObserver{id: "obs1"}},
			wantLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, o := range tt.observers {
				tt.pub.Register(o)
			}
			assert.Len(t, tt.pub.observers, tt.wantLen)
		})
	}
}

func TestPub_Notify(t *testing.T) {
	obs1 := &mockObserver{id: "obs1"}
	obs2 := &mockObserver{id: "obs2"}
	pub := &Pub{}
	pub.Register(obs1)
	pub.Register(obs2)

	metrics := []string{"cpu", "mem"}
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	pub.Notify(req, metrics)

	assert.Equal(t, metrics, obs1.lastMetrics)
	assert.Equal(t, metrics, obs2.lastMetrics)
	assert.Equal(t, "192.168.1.1", obs1.lastIP)
	assert.Equal(t, "192.168.1.1", obs2.lastIP)
}

func TestPub_Notify_NoObservers(t *testing.T) {
	pub := &Pub{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5678"

	assert.NotPanics(t, func() {
		pub.Notify(req, nil)
	})
}

func TestPub_getClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		wantIP     string
	}{
		{
			name:       "x-forwarded-for first ip",
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.1, 70.41.3.18, 150.172.238.178",
			wantIP:     "203.0.113.1",
		},
		{
			name:       "x-forwarded-for single ip",
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.1",
			wantIP:     "203.0.113.1",
		},
		{
			name:       "x-real-ip",
			remoteAddr: "10.0.0.1:1234",
			xri:        "203.0.113.2",
			wantIP:     "203.0.113.2",
		},
		{
			name:       "x-forwarded-for takes precedence over x-real-ip",
			remoteAddr: "10.0.0.1:1234",
			xff:        "203.0.113.1",
			xri:        "203.0.113.2",
			wantIP:     "203.0.113.1",
		},
		{
			name:       "remote addr with port",
			remoteAddr: "192.168.1.1:8080",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "remote addr without port",
			remoteAddr: "192.168.1.1",
			wantIP:     "192.168.1.1",
		},
		{
			name:       "ipv6 remote addr",
			remoteAddr: "[::1]:8080",
			wantIP:     "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			pub := &Pub{}
			got := pub.getClientIP(req)
			assert.Equal(t, tt.wantIP, got)
		})
	}
}
