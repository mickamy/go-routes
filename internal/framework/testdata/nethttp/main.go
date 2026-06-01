package main

import (
	"net/http"

	"example.com/nethttpapp/handlers"
)

func main() {
	http.HandleFunc("GET /", handlers.Index)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /posts", handlers.List)
	mux.HandleFunc("POST /posts", handlers.Create)
	mux.HandleFunc("/legacy", handlers.Index)
	mux.HandleFunc("DELETE /posts/{id}", func(w http.ResponseWriter, r *http.Request) {})

	_ = http.ListenAndServe(":8080", mux)
}
