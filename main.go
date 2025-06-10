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

	// Admin
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)
	mux.HandleFunc("GET /api/healthz", cfg.HandlerReadiness)

	// Auth
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.Login(w, r, cfg)
	})
	mux.HandleFunc("POST /api/refresh", func(w http.ResponseWriter, r *http.Request) {
		handlers.RefreshAccessToken(w, r, cfg)
	})
	mux.HandleFunc("POST /api/revoke", func(w http.ResponseWriter, r *http.Request) {
		handlers.RevokeRefreshToken(w, r, cfg)
	})

	// Users
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateUser(w, r, cfg)
	})
	mux.HandleFunc("PUT /api/users", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateUser(w, r, cfg)
	})

	// Chirps
	mux.HandleFunc("POST /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateChirp(w, r, cfg)
	})
	mux.HandleFunc("GET /api/chirps", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetAllChirps(w, r, cfg)
	})
	mux.HandleFunc("GET /api/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		chirpID, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			api.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		handlers.GetChirp(w, r, cfg, chirpID)
	})
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		chirpID, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			api.RespondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		handlers.DeleteChirp(w, r, cfg, chirpID)
	})

	// Hooks
	mux.HandleFunc("POST /api/polka/webhooks", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpgradeToChirpyRed(w, r, cfg)
	})

	server := &http.Server{Addr: ":8080", Handler: mux}
	server.ListenAndServe()
}
