package repository

type MemStorage struct {
	data map[string]any
}

func (m MemStorage) AddGauge(name string, value float64) {
	// новое значение должно замещать предыдущее
	m.data[name] = value
}

func (m MemStorage) CheckGaugeType(name string) bool {
	val, ok := m.data[name]
	_, okType := val.(float64)
	if ok && okType {
		return true
	}
	return false
}

func (m MemStorage) AddCounter(name string, value int64) {
	// новое значение должно добавляться к предыдущему
	oldValue, ok := m.data[name].(int64)
	if ok {
		m.data[name] = oldValue + value
	} else {
		m.data[name] = value
	}
}

func (m MemStorage) CheckCounterType(name string) bool {
	val, ok := m.data[name]
	_, okType := val.(int64)
	if ok && okType {
		return true
	}
	return false
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		data: make(map[string]any),
	}
}
