package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	gzip "github.com/postman17/metrics/internal/gzip"
	handlers "github.com/postman17/metrics/internal/handler"
	log "github.com/postman17/metrics/internal/logger"
	repo "github.com/postman17/metrics/internal/repository"
)

func main() {
	log.InitializeLogger("INFO")
	config := parseFlags()

	storeNotSync := *config.StoreInterval > 0
	memory := repo.NewMemStorage(storeNotSync, config.FileStoragePath)
	if *config.Restore {
		memory.LoadFromFile()
	}

	if storeNotSync {
		go func() {
			ticker := time.NewTicker(time.Duration(*config.StoreInterval) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				if err := memory.SaveToFile(); err != nil {
					slog.Error(
						"memory save to file failed", "err", err,
					)
				}
			}
		}()
	}

	r := chi.NewRouter()
	r.Use(log.WithLogging)
	r.Use(gzip.GZIPMiddleware)

	r.Get("/", handlers.GetMainPage(memory))
	r.Post("/update/", handlers.UpdateMetric(memory))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricPage(memory))
	r.Post("/value/", handlers.GetMetricValue(memory))
	r.Get("/value/{type}/{name}", handlers.GetMetricValuePage(memory))

	err := http.ListenAndServe(config.RunAddr, r)
	if err != nil {
		panic(err)
	}
}
