package repository

import (
	"sync"

	models "github.com/postman17/metrics/internal/model"
)

type MemStorage struct {
	data map[string]any
	mu   sync.Mutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]any),
	}
}

func NewMemStorageWithData(data map[string]any) *MemStorage {
	m := NewMemStorage()
	m.mu.Lock()
	for k, v := range data {
		m.data[k] = v
	}
	m.mu.Unlock()
	return m
}

func (m *MemStorage) AddGauge(name string, value float64) {
	m.mu.Lock()
	m.data[name] = value
	m.mu.Unlock()
}

func (m *MemStorage) CheckGaugeType(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[name]
	_, okType := val.(float64)
	return ok && okType
}

func (m *MemStorage) AddCounter(name string, value int64) {
	m.mu.Lock()
	oldValue, ok := m.data[name].(int64)
	if ok {
		m.data[name] = oldValue + value
	} else {
		m.data[name] = value
	}
	m.mu.Unlock()
}

func (m *MemStorage) CheckCounterType(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[name]
	_, okType := val.(int64)
	return ok && okType
}

func (m *MemStorage) GetTypeValue(name string) any {
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[name]
	if !ok {
		return nil
	}
	return val
}

func (m *MemStorage) GetAll() map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]any, len(m.data))
	for k, v := range m.data {
		result[k] = v
	}
	return result
}

func (m *MemStorage) AddBatch(data models.MetricsList) error {
	if err := validateBatch(data); err != nil {
		return err
	}

	m.mu.Lock()
	for _, metric := range data {
		switch metric.MType {
		case models.Gauge:
			m.data[metric.ID] = *metric.Value
		case models.Counter:
			oldValue, ok := m.data[metric.ID].(int64)
			if ok {
				m.data[metric.ID] = oldValue + *metric.Delta
			} else {
				m.data[metric.ID] = *metric.Delta
			}
		}
	}
	m.mu.Unlock()
	return nil
}

func (m *MemStorage) loadFromList(list models.MetricsList) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, metric := range list {
		switch metric.MType {
		case models.Counter:
			if metric.Delta != nil {
				m.data[metric.ID] = *metric.Delta
			}
		case models.Gauge:
			if metric.Value != nil {
				m.data[metric.ID] = *metric.Value
			}
		}
	}
}

func metricsListFromData(data map[string]any) models.MetricsList {
	list := make(models.MetricsList, 0, len(data))
	for id, v := range data {
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
	return list
}
