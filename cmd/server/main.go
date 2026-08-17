// Сервер мониторинга принимает метрики от агентов, хранит их
// (in-memory, file или PostgreSQL) и отдаёт по HTTP.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	audit "github.com/postman17/metrics/internal/audit"
	dbconfig "github.com/postman17/metrics/internal/config/db"
	gzip "github.com/postman17/metrics/internal/gzip"
	handlers "github.com/postman17/metrics/internal/handler"
	log "github.com/postman17/metrics/internal/logger"
	repo "github.com/postman17/metrics/internal/repository"
	sha256mw "github.com/postman17/metrics/internal/sha256"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	fmt.Println("Build version:", buildVersion)
	if buildDate == "" {
		buildDate = "N/A"
	}
	fmt.Println("Build date:", buildDate)
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	fmt.Println("Build commit:", buildCommit)
	if err := run(); err != nil {
		fail(1)
	}
}

func fail(code int) {
	os.Exit(code)
}

func run() error {
	if err := log.InitializeLogger("INFO"); err != nil {
		slog.Error("failed to initialize logger", "err", err)
		return err
	}
	config := parseFlags()

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	storage, db, err := newMetricsRepository(appCtx, config)
	if err != nil {
		slog.Error("failed to initialize storage", "err", err)
		return err
	}
	if db != nil {
		defer func() { _ = db.Close() }()
	}

	if persistent, ok := storage.(repo.PersistentRepository); ok && !(*config.StoreInterval == 0) {
		go runPeriodicSave(appCtx, persistent, time.Duration(*config.StoreInterval)*time.Second)
	}

	pub := audit.Pub{}
	if config.AuditFile != "" {
		pub.Register(&audit.FileSubscriber{ID: "file", FilePath: config.AuditFile})
	}
	if config.AuditURL != "" {
		pub.Register(audit.NewURLSubscriber("url", config.AuditURL))
	}

	r := chi.NewRouter()
	r.Use(log.WithLogging)
	r.Use(gzip.GZIPMiddleware)
	r.Use(sha256mw.Middleware(config.Key))

	if db != nil {
		r.Get("/ping", handlers.Ping(db))
	}

	r.Get("/", handlers.GetMainPage(storage))
	r.Post("/update/", handlers.UpdateMetric(storage))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricPage(storage))
	r.Post("/value/", handlers.GetMetricValue(storage))
	r.Get("/value/{type}/{name}", handlers.GetMetricValuePage(storage))
	r.Post("/updates/", handlers.UpdatesMetric(storage, &pub))

	srv := &http.Server{Addr: config.RunAddr, Handler: r}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start http server", "err", err)
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
	case err := <-serverErr:
		if err != nil {
			return err
		}
	}

	if persistent, ok := storage.(repo.PersistentRepository); ok {
		if err := persistent.Save(); err != nil {
			slog.Error("storage save failed", "err", err)
		}
	}

	appCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "err", err)
	}

	return nil
}

// newMetricsRepository создаёт хранилище метрик на основе конфигурации:
// DBStorage при наличии DSN, FileStorage при указании пути к файлу, иначе MemStorage.
func newMetricsRepository(ctx context.Context, config Config) (repo.MetricsRepository, *sql.DB, error) {
	if config.DatabaseDSN != "" {
		db, err := dbconfig.Open(ctx, config.DatabaseDSN)
		if err != nil {
			return nil, nil, err
		}

		if err := runMigrations(db); err != nil {
			if db != nil {
				_ = db.Close()
			}
			return nil, nil, err
		}
		slog.Info("database migrations applied")

		return repo.NewDBStorage(ctx, db), db, nil
	}

	if config.FileStoragePath != "" {
		storeSync := *config.StoreInterval == 0
		storage, err := repo.NewFileStorage(config.FileStoragePath, storeSync, *config.Restore)
		if err != nil {
			return nil, nil, err
		}
		return storage, nil, nil
	}

	return repo.NewMemStorage(), nil, nil
}

// runPeriodicSave периодически вызывает Save() хранилища с заданным интервалом,
// пока контекст не будет отменён.
func runPeriodicSave(ctx context.Context, storage repo.PersistentRepository, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := storage.Save(); err != nil {
				slog.Error("storage save failed", "err", err)
			}
		}
	}
}
