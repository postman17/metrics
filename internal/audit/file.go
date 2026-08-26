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
	if mkdirErr := os.MkdirAll(dir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, mkdirErr)
	}

	file, err := os.OpenFile(s.FilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open/create file %s: %w", s.FilePath, err)
	}
	defer func() { _ = file.Close() }()

	if _, writeErr := file.Write(data); writeErr != nil {
		return fmt.Errorf("failed to write metrics data: %w", writeErr)
	}

	if _, nlErr := file.Write([]byte("\n")); nlErr != nil {
		return fmt.Errorf("failed to write newline separator: %w", nlErr)
	}

	slog.Info("Data appended to file", "path", s.FilePath)
	return nil
}
