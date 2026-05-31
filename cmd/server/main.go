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
	_ "github.com/jackc/pgx/v5/stdlib"
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

	DB := &sql.DB{}
	var err error
	withDB := config.Database_DSN != ""
	if withDB {
		DB, err = sql.Open("pgx", config.Database_DSN)
		if err != nil {
			slog.Error("failed db configuration", "err", err)
			os.Exit(1)
		}
		defer DB.Close()

		if err := runMigrations(DB); err != nil {
			slog.Error("failed to apply migrations", "err", err)
			os.Exit(1)
		}
		slog.Info("database migrations applied")
	}

	storeSync := *config.StoreInterval == 0
	memory := repo.NewMemStorage(appCtx, storeSync, config.FileStoragePath, *config.Restore, DB, withDB)

	if !storeSync && config.FileStoragePath != "" {
		go func() {
			ticker := time.NewTicker(time.Duration(*config.StoreInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-appCtx.Done():
					return
				case <-ticker.C:
					if err := memory.SaveToFile(); err != nil {
						slog.Error(
							"memory save to file failed", "err", err,
						)
					}
				}
			}
		}()
	}

	r := chi.NewRouter()
	r.Use(log.WithLogging)
	r.Use(gzip.GZIPMiddleware)

	if withDB {
		r.Get("/ping", handlers.Ping(DB))
	}

	r.Get("/", handlers.GetMainPage(memory))
	r.Post("/update/", handlers.UpdateMetric(memory))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricPage(memory))
	r.Post("/value/", handlers.GetMetricValue(memory))
	r.Get("/value/{type}/{name}", handlers.GetMetricValuePage(memory))
	r.Post("/updates/", handlers.UpdatesMetric(memory))

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

	if config.FileStoragePath != "" {
		if err := memory.SaveToFile(); err != nil {
			slog.Error("memory save to file failed", "err", err)
		}
	}

	appCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "err", err)
	}
}
