package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	handlers "github.com/postman17/metrics/internal/handler"
	log "github.com/postman17/metrics/internal/logger"
	repo "github.com/postman17/metrics/internal/repository"
)

func main() {
	log.InitializeLogger("INFO")
	config := parseFlags()

	memory := repo.NewMemStorage()

	r := chi.NewRouter()
	r.Use(log.WithLogging)

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
