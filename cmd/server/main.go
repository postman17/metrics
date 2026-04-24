package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	handlers "github.com/postman17/metrics/internal/handler"
	repo "github.com/postman17/metrics/internal/repository"
)

func main() {
	memory := repo.NewMemStorage()

	r := chi.NewRouter()

	r.Get("/", handlers.GetMainPage(memory))
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricPage(memory))
	r.Get("/value/{type}/{name}", handlers.GetMetricValuePage(memory))

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
