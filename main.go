package main

import (
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/portbound/bootdev-httpserver/api"
	"github.com/portbound/bootdev-httpserver/api/handlers"
)

func main() {
	cfg, err := api.NewConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.MiddlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)

	mux.HandleFunc("GET /api/healthz", cfg.HandlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlers.ValidateChirp)
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUser(w, r, cfg)
	})

	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
