package main

import (
	"net/http"

	"example.com/nethttpapp/handlers"
)

const (
	apiBase     = "/api"
	healthRoute = "GET /healthz"
)

func main() {
	http.HandleFunc("GET /", handlers.Index)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts", handlers.List)
	mux.HandleFunc("POST /posts", handlers.Create)
	mux.HandleFunc("/legacy", handlers.Index)
	mux.HandleFunc("DELETE /posts/{id}", func(w http.ResponseWriter, r *http.Request) {})

	mux.HandleFunc(healthRoute, handlers.Index)
	mux.HandleFunc("GET "+apiBase+"/comments", handlers.List)

	_ = http.ListenAndServe(":8080", mux)
}
