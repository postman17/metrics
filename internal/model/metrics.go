// Package models определяет структуры данных для представления метрик и запросов
// к сервису мониторинга. Включает модели для counter и gauge метрик,
// а также структуры для аудита.
//
//go:generate go tool easyjson -all metrics.go
package models

// Counter — тип метрики-счётчика, значение которого может только увеличиваться.
const Counter = "counter"

// Gauge — тип метрики-измерения, значение которого может произвольно меняться.
const Gauge = "gauge"

// NOTE: Не усложняем пример, вводя иерархическую вложенность структур.
// Органичиваясь плоской моделью.
// Delta и Value объявлены через указатели,
// что бы отличать значение "0", от не заданного значения
// и соответственно не кодировать в структуру.
//
// Metrics представляет одну метрику сервера мониторинга.
// Поля Delta и Value — указатели, чтобы отличать нулевое значение
// от отсутствующего и не сериализовать пустое значение в JSON.
//
//easyjson:json
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// GetMetricRequest описывает JSON-запрос на получение значения метрики по ID и типу.
//
//easyjson:json
type GetMetricRequest struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}

// MetricsList — срез метрик, используется для пакетного обновления.
//
//easyjson:json
type MetricsList []Metrics

// AuditMetrics описывает запись аудита: timestamp, список обновлённых метрик и IP-адрес клиента.
//
//easyjson:json
type AuditMetrics struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}
