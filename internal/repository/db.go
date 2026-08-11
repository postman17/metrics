package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	models "github.com/postman17/metrics/internal/model"
)

// DBStorage — хранилище метрик в PostgreSQL.
type DBStorage struct {
	db  *sql.DB
	ctx context.Context
}

// NewDBStorage создаёт хранилище, использующее указанное подключение к PostgreSQL.
func NewDBStorage(ctx context.Context, db *sql.DB) *DBStorage {
	return &DBStorage{
		db:  db,
		ctx: ctx,
	}
}

// AddGauge вставляет или обновляет gauge-метрику в БД (upsert).
func (d *DBStorage) AddGauge(name string, value float64) {
	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	_, err := d.db.ExecContext(ctx, `
		INSERT INTO metrics (name, m_type, value, delta) VALUES ($1, 'gauge', $2, 0)
		ON CONFLICT (name) DO UPDATE SET
			m_type = 'gauge',
			value = EXCLUDED.value,
			delta = 0
	`, name, value)
	if err != nil {
		slog.Error("cant save gauge to db", "err", err)
	}
}

// AddCounter вставляет или увеличивает counter-метрику в БД (upsert с накоплением delta).
func (d *DBStorage) AddCounter(name string, value int64) {
	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	_, err := d.db.ExecContext(ctx, `
		INSERT INTO metrics (name, m_type, delta, value) VALUES ($1, 'counter', $2, 0)
		ON CONFLICT (name) DO UPDATE SET
			m_type = 'counter',
			delta = CASE WHEN metrics.m_type = 'counter' THEN metrics.delta + EXCLUDED.delta ELSE EXCLUDED.delta END,
			value = 0
	`, name, value)
	if err != nil {
		slog.Error("cant save counter to db", "err", err)
	}
}

// CheckGaugeType возвращает true, если метрика name существует и имеет тип gauge.
func (d *DBStorage) CheckGaugeType(name string) bool {
	mType, ok := d.getMetricType(name)
	return ok && mType == models.Gauge
}

// CheckCounterType возвращает true, если метрика name существует и имеет тип counter.
func (d *DBStorage) CheckCounterType(name string) bool {
	mType, ok := d.getMetricType(name)
	return ok && mType == models.Counter
}

// GetTypeValue возвращает значение метрики name из БД или nil, если она не найдена.
func (d *DBStorage) GetTypeValue(name string) any {
	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	var (
		mType string
		value sql.NullFloat64
		delta sql.NullInt64
	)
	err := d.db.QueryRowContext(
		ctx, "SELECT m_type, value, delta FROM metrics WHERE name = $1", name,
	).Scan(&mType, &value, &delta)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		slog.Error("cant read metric from db", "err", err)
		return nil
	}

	switch mType {
	case models.Gauge:
		if value.Valid {
			return value.Float64
		}
	case models.Counter:
		if delta.Valid {
			return delta.Int64
		}
	}
	return nil
}

// GetAll возвращает все метрики из БД в виде map[name]value.
func (d *DBStorage) GetAll() map[string]any {
	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	rows, err := d.db.QueryContext(ctx, "SELECT name, m_type, value, delta FROM metrics")
	if err != nil {
		slog.Error("cant read metrics from db", "err", err)
		return map[string]any{}
	}
	defer rows.Close()

	result := make(map[string]any)
	for rows.Next() {
		var (
			name  string
			mType string
			value sql.NullFloat64
			delta sql.NullInt64
		)
		if err := rows.Scan(&name, &mType, &value, &delta); err != nil {
			slog.Error("cant scan metric row", "err", err)
			continue
		}

		switch mType {
		case models.Gauge:
			if value.Valid {
				result[name] = value.Float64
			}
		case models.Counter:
			if delta.Valid {
				result[name] = delta.Int64
			}
		}
	}
	if err := rows.Err(); err != nil {
		slog.Error("rows iteration error", "err", err)
	}
	return result
}

// AddBatch вставляет или обновляет срез метрик в одной транзакции.
func (d *DBStorage) AddBatch(data models.MetricsList) error {
	if err := validateBatch(data); err != nil {
		return err
	}

	txCtx, cancel := context.WithTimeout(d.ctx, 30*time.Second)
	defer cancel()

	tx, err := d.db.BeginTx(txCtx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmtGauge, err := tx.PrepareContext(txCtx, `
		INSERT INTO metrics (name, m_type, value, delta) VALUES ($1, 'gauge', $2, 0)
		ON CONFLICT (name) DO UPDATE SET
			m_type = 'gauge',
			value = EXCLUDED.value,
			delta = 0
	`)
	if err != nil {
		return err
	}
	defer stmtGauge.Close()

	stmtCounter, err := tx.PrepareContext(txCtx, `
		INSERT INTO metrics (name, m_type, delta, value) VALUES ($1, 'counter', $2, 0)
		ON CONFLICT (name) DO UPDATE SET
			m_type = 'counter',
			delta = CASE WHEN metrics.m_type = 'counter' THEN metrics.delta + EXCLUDED.delta ELSE EXCLUDED.delta END,
			value = 0
	`)
	if err != nil {
		return err
	}
	defer stmtCounter.Close()

	for _, metric := range data {
		switch metric.MType {
		case models.Gauge:
			_, err = stmtGauge.ExecContext(txCtx, metric.ID, *metric.Value)
		case models.Counter:
			_, err = stmtCounter.ExecContext(txCtx, metric.ID, *metric.Delta)
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (d *DBStorage) getMetricType(name string) (string, bool) {
	ctx, cancel := context.WithTimeout(d.ctx, 5*time.Second)
	defer cancel()

	var mType string
	err := d.db.QueryRowContext(ctx, "SELECT m_type FROM metrics WHERE name = $1", name).Scan(&mType)
	if err == sql.ErrNoRows {
		return "", false
	}
	if err != nil {
		slog.Error("cant read metric type from db", "err", err)
		return "", false
	}
	return mType, true
}
