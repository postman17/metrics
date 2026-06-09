package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	dbconfig "github.com/postman17/metrics/internal/config/db"
	gzip "github.com/postman17/metrics/internal/gzip"
	handlers "github.com/postman17/metrics/internal/handler"
	log "github.com/postman17/metrics/internal/logger"
	repo "github.com/postman17/metrics/internal/repository"
)

func main() {
	if err := log.InitializeLogger("INFO"); err != nil {
		slog.Error("failed to initialize logger", "err", err)
		os.Exit(1)
	}
	config := parseFlags()

	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	storage, db, err := newMetricsRepository(appCtx, config)
	if err != nil {
		slog.Error("failed to initialize storage", "err", err)
		os.Exit(1)
	}
	if db != nil {
		defer db.Close()
	}

	if persistent, ok := storage.(repo.PersistentRepository); ok && !(*config.StoreInterval == 0) {
		go runPeriodicSave(appCtx, persistent, time.Duration(*config.StoreInterval)*time.Second)
	}

	r := chi.NewRouter()
	r.Use(log.WithLogging)
	r.Use(gzip.GZIPMiddleware)

	if db != nil {
		r.Get("/ping", handlers.Ping(db))
	}

	r.Get("/", handlers.GetMainPage(storage))
	r.Post("/update/", handlers.UpdateMetric(storage))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricPage(storage))
	r.Post("/value/", handlers.GetMetricValue(storage))
	r.Get("/value/{type}/{name}", handlers.GetMetricValuePage(storage))
	r.Post("/updates/", handlers.UpdatesMetric(storage))

	srv := &http.Server{Addr: config.RunAddr, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start http server", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

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
}

func newMetricsRepository(ctx context.Context, config Config) (repo.MetricsRepository, *sql.DB, error) {
	if config.Database_DSN != "" {
		db, err := dbconfig.Open(ctx, config.Database_DSN)
		if err != nil {
			return nil, nil, err
		}

		if err := runMigrations(db); err != nil {
			db.Close()
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
