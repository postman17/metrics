// Package audit реализует паттерн «наблюдатель» для аудита операций с метриками.
// Pub рассылает уведомления зарегистрированным подписчикам (FileSubscriber, URLSubscriber).
package audit

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
)

type observer interface {
	send(string, []string) error
	getID() string
}

// Pub — издатель уведомлений об обновлении метрик.
type Pub struct {
	observers map[string]observer
}

// Register добавляет подписчика в издатель.
func (e *Pub) Register(o observer) {
	if e.observers == nil {
		e.observers = make(map[string]observer)
	}
	e.observers[o.getID()] = o
}

// Notify рассылает уведомление всем подписчикам с IP-адресом клиента и списком обновлённых метрик.
func (e *Pub) Notify(r *http.Request, metrics []string) {
	for _, observer := range e.observers {
		go func() {
			if err := observer.send(e.getClientIP(r), metrics); err != nil {
				slog.Error("audit observer send failed", "id", observer.getID(), "err", err)
			}
		}()
	}
}

func (e *Pub) getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if parts := strings.Split(xff, ","); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
