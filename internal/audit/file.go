package audit

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
)

// FileSubscriber — подписчик, записывающий аудит-события в файл построчно в JSON.
type FileSubscriber struct {
	ID       string
	FilePath string
	mu       sync.Mutex
}

func (s *FileSubscriber) getID() string {
	return s.ID
}

func (s *FileSubscriber) send(ip string, metrics []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	timestamp := time.Now().Unix()
	met := models.AuditMetrics{TS: timestamp, IPAddress: ip, Metrics: metrics}
	data, err := easyjson.Marshal(met)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics via easyjson: %w", err)
	}

	dir := filepath.Dir(s.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.OpenFile(s.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open/create file %s: %w", s.FilePath, err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write metrics data: %w", err)
	}

	if _, err := file.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write newline separator: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file to disk: %w", err)
	}

	slog.Info("Data safely appended to file", "path", s.FilePath)
	return nil
}
