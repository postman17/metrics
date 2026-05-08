package main

import (
	"net/http"

	handlers "github.com/postman17/metrics/internal/handler"
	repo "github.com/postman17/metrics/internal/repository"
)

func main() {
	memory := repo.NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/{type}/{name}/{value}`, handlers.UpdateMetricPage(memory))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
