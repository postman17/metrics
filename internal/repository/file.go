package repository

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/mailru/easyjson"
	models "github.com/postman17/metrics/internal/model"
)

// FileStorage — хранилище метрик с персистентностью на файловую систему.
// Делегирует хранение в памяти объекту MemStorage и при необходимости
// записывает/читает данные из файла.
type FileStorage struct {
	mem       *MemStorage
	filePath  string
	storeSync bool
}

// NewFileStorage создаёт файловое хранилище.
// filePath — путь к файлу для персистентности;
// storeSync — если true, данные сохраняются после каждой операции;
// restore — если true, данные загружаются из файла при старте.
func NewFileStorage(filePath string, storeSync bool, restore bool) (*FileStorage, error) {
	if filePath == "" {
		return nil, errEmptyFilePath()
	}

	storage := &FileStorage{
		mem:       NewMemStorage(),
		filePath:  filePath,
		storeSync: storeSync,
	}

	if restore {
		if err := storage.load(); err != nil {
			return nil, err
		}
	}

	return storage, nil
}

// AddGauge добавляет или перезаписывает gauge-метрику и при storeSync сохраняет файл.
func (f *FileStorage) AddGauge(name string, value float64) {
	f.mem.AddGauge(name, value)
	f.persistIfSync()
}

// AddCounter увеличивает counter-метрику на value и при storeSync сохраняет файл.
func (f *FileStorage) AddCounter(name string, value int64) {
	f.mem.AddCounter(name, value)
	f.persistIfSync()
}

// CheckGaugeType возвращает true, если метрика name имеет тип gauge.
func (f *FileStorage) CheckGaugeType(name string) bool {
	return f.mem.CheckGaugeType(name)
}

// CheckCounterType возвращает true, если метрика name имеет тип counter.
func (f *FileStorage) CheckCounterType(name string) bool {
	return f.mem.CheckCounterType(name)
}

// GetTypeValue возвращает текущее значение метрики name или nil, если она не найдена.
func (f *FileStorage) GetTypeValue(name string) any {
	return f.mem.GetTypeValue(name)
}

// GetAll возвращает копию всех метрик в виде map[name]value.
func (f *FileStorage) GetAll() map[string]any {
	return f.mem.GetAll()
}

// AddBatch добавляет срез метрик и при storeSync сохраняет файл.
func (f *FileStorage) AddBatch(data models.MetricsList) error {
	if err := f.mem.AddBatch(data); err != nil {
		return err
	}
	f.persistIfSync()
	return nil
}

// Save явно сохраняет текущее состояние хранилища в файл.
func (f *FileStorage) Save() error {
	return f.save()
}

func (f *FileStorage) persistIfSync() {
	if !f.storeSync {
		return
	}
	if err := f.save(); err != nil {
		slog.Error("file storage save failed", "err", err)
	}
}

func (f *FileStorage) load() error {
	data, err := os.ReadFile(f.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var list models.MetricsList
	if err := easyjson.Unmarshal(data, &list); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	f.mem.loadFromList(list)
	slog.Info("Data load from file")
	return nil
}

func (f *FileStorage) save() error {
	list := metricsListFromData(f.mem.GetAll())

	data, err := easyjson.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics via easyjson: %w", err)
	}

	dir := filepath.Dir(f.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tmpPath := f.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpPath, f.filePath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file to destination: %w", err)
	}

	slog.Info("Data saved to file")
	return nil
}

func errEmptyFilePath() error {
	return fmt.Errorf("path to storage file is empty")
}
