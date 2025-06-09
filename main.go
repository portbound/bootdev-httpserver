package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
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

	cfg.JWT = os.Getenv("JWT")

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.MiddlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)
	mux.HandleFunc("GET /api/healthz", cfg.HandlerReadiness)

	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateChirp(w, r, cfg)
	})
	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetAllChirps(w, r, cfg)
	})
	mux.HandleFunc("GET /api/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		chirpId, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			return
		}
		handlers.GetChirp(w, r, cfg, chirpId)
	})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUser(w, r, cfg)
	})
	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateUser(w, r, cfg)
	})
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.Login(w, r, cfg)
	})
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlers.RefreshAccessToken(w, r, cfg)
	})
	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, r *http.Request) {
		handlers.RevokeRefreshToken(w, r, cfg)
	})
	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
