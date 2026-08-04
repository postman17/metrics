// Package db содержит вспомогательные функции для подключения к PostgreSQL,
// включая повторные попытки при устранимых ошибках соединения.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// connectRetryDelays — задержки между повторными попытками подключения к БД.
var connectRetryDelays = []time.Duration{
	1 * time.Second,
	3 * time.Second,
	5 * time.Second,
}

// Open открывает пул подключений к PostgreSQL и проверяет доступность БД с повторами.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= len(connectRetryDelays); attempt++ {
		if attempt > 0 {
			delay := connectRetryDelays[attempt-1]
			slog.Warn(
				"database connect failed, retrying",
				"err", lastErr,
				"attempt", attempt,
				"delay", delay,
			)

			select {
			case <-ctx.Done():
				_ = db.Close()
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		lastErr = db.PingContext(pingCtx)
		cancel()

		if lastErr == nil {
			return db, nil
		}
		if !isRetryableConnectError(lastErr) {
			_ = db.Close()
			return nil, fmt.Errorf("connect database: %w", lastErr)
		}
	}

	_ = db.Close()
	return nil, fmt.Errorf("connect database: %w", lastErr)
}

func isRetryableConnectError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "08")
	}

	return false
}
