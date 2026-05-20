package repository

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
)

type MemStorage struct {
	Data         map[string]any
	StoreNotSync bool
	FilePath     string
	mu           sync.RWMutex
}

func (m *MemStorage) AddGauge(name string, value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// новое значение должно замещать предыдущее
	m.Data[name] = value
}

func (m *MemStorage) CheckGaugeType(name string) bool {
	val, ok := m.Data[name]
	_, okType := val.(float64)
	if ok && okType {
		return true
	}
	return false
}

func (m *MemStorage) AddCounter(name string, value int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// новое значение должно добавляться к предыдущему
	oldValue, ok := m.Data[name].(int64)
	if ok {
		m.Data[name] = oldValue + value
	} else {
		m.Data[name] = value
	}
}

func (m *MemStorage) CheckCounterType(name string) bool {
	val, ok := m.Data[name]
	_, okType := val.(int64)
	if ok && okType {
		return true
	}
	return false
}

func (m *MemStorage) GetTypeValue(name string) any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.Data[name]
	if !ok {
		return nil
	}
	return val
}

func (m *MemStorage) LoadFromFile() error {
	if m.FilePath == "" {
		return fmt.Errorf("path to storage file is empty")
	}

	if _, err := os.Stat(m.FilePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(m.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var list models.MetricsList
	if err := easyjson.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	for _, metric := range list {
		switch metric.MType {
		case models.Counter:
			if metric.Delta != nil {
				m.Data[metric.ID] = *metric.Delta
			}
		case models.Gauge:
			if metric.Value != nil {
				m.Data[metric.ID] = *metric.Value
			}
		}
	}
	slog.Info("Data load from file")
	return nil
}

func (m *MemStorage) SaveToFile() error {
	m.mu.RLock()
	list := make(models.MetricsList, 0, len(m.Data))

	for id, v := range m.Data {
		switch val := v.(type) {
		case float64:
			value := val
			list = append(list, models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &value,
			})
		case int64:
			delta := val
			list = append(list, models.Metrics{
				ID:    id,
				MType: models.Counter,
				Delta: &delta,
			})
		}
	}
	m.mu.RUnlock()

	data, err := easyjson.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics via easyjson: %w", err)
	}

	dir := filepath.Dir(m.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpPath := m.FilePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, m.FilePath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file to destination: %w", err)
	}

	slog.Info("Data saved to file")

	return nil
}

func NewMemStorage(storeNotSync bool, filePath string) *MemStorage {
	return &MemStorage{
		Data:         make(map[string]any),
		StoreNotSync: storeNotSync,
		FilePath:     filePath,
	}
}
