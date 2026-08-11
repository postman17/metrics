// Package repository определяет интерфейсы и реализации хранилища метрик:
// MetricsRepository — базовый CRUD, PersistentRepository — с поддержкой сохранения на диск,
// а также in-memory, file-based и PostgreSQL реализации.
package repository

import (
	"fmt"

	models "github.com/postman17/metrics/internal/model"
)

// MetricsRepository определяет интерфейс хранилища метрик.
// Реализации: MemStorage, FileStorage, DBStorage.
type MetricsRepository interface {
	AddGauge(name string, value float64)
	AddCounter(name string, value int64)
	CheckGaugeType(name string) bool
	CheckCounterType(name string) bool
	GetTypeValue(name string) any
	GetAll() map[string]any
	AddBatch(data models.MetricsList) error
}

// PersistentRepository расширяет MetricsRepository возможностью явного
// сохранения состояния (например, на диск).
type PersistentRepository interface {
	MetricsRepository
	Save() error
}

func validateBatch(data models.MetricsList) error {
	for _, metric := range data {
		if metric.ID == "" {
			return fmt.Errorf("metric id is empty")
		}
		if metric.MType != models.Gauge && metric.MType != models.Counter {
			return fmt.Errorf("invalid metric type: %s", metric.MType)
		}
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("gauge value is nil for metric %s", metric.ID)
			}
		case models.Counter:
			if metric.Delta == nil {
				return fmt.Errorf("counter delta is nil for metric %s", metric.ID)
			}
		}
	}
	return nil
}

var (
	_ MetricsRepository    = (*MemStorage)(nil)
	_ MetricsRepository    = (*FileStorage)(nil)
	_ MetricsRepository    = (*DBStorage)(nil)
	_ PersistentRepository = (*FileStorage)(nil)
)
