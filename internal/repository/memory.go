package repository

type MemStorage struct {
	Data map[string]any
}

func (m MemStorage) AddGauge(name string, value float64) {
	// новое значение должно замещать предыдущее
	m.Data[name] = value
}

func (m MemStorage) CheckGaugeType(name string) bool {
	val, ok := m.Data[name]
	_, okType := val.(float64)
	if ok && okType {
		return true
	}
	return false
}

func (m MemStorage) AddCounter(name string, value int64) {
	// новое значение должно добавляться к предыдущему
	oldValue, ok := m.Data[name].(int64)
	if ok {
		m.Data[name] = oldValue + value
	} else {
		m.Data[name] = value
	}
}

func (m MemStorage) CheckCounterType(name string) bool {
	val, ok := m.Data[name]
	_, okType := val.(int64)
	if ok && okType {
		return true
	}
	return false
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Data: make(map[string]any),
	}
}
