package audit

import (
	"net"
	"net/http"
	"strings"
)

type observer interface {
	send(string, []string) error
	getID() string
}

type Pub struct {
	observers map[string]observer
}

func (e *Pub) Register(o observer) {
	if e.observers == nil {
		e.observers = make(map[string]observer)
	}
	e.observers[o.getID()] = o
}

func (e *Pub) Notify(r *http.Request, metrics []string) {
	for _, observer := range e.observers {
		observer.send(e.getClientIP(r), metrics)
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
