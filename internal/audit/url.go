package audit

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
)

// URLSubscriber — подписчик, отправляющий аудит-события HTTP POST на внешний URL.
type URLSubscriber struct {
	ID  string
	URL string
	client *http.Client
}

// NewURLSubscriber создаёт URLSubscriber с HTTP-клиентом, имеющим таймаут.
func NewURLSubscriber(id, url string) *URLSubscriber {
	return &URLSubscriber{
		ID:  id,
		URL: url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *URLSubscriber) getID() string {
	return s.ID
}

func (s *URLSubscriber) send(ip string, metrics []string) error {
	timestamp := time.Now().Unix()
	met := models.AuditMetrics{TS: timestamp, IPAddress: ip, Metrics: metrics}
	data, err := easyjson.Marshal(met)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics via easyjson: %w", err)
	}

	body := bytes.NewBuffer(data)

	resp, err := s.client.Post(s.URL, "application/json", body)
	if err != nil {
		return fmt.Errorf("failed to send POST request to %s: %w", s.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("bad status code from subscriber %s: %d", s.URL, resp.StatusCode)
	}

	return nil
}
