package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	"github.com/portbound/bootdev-httpserver/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(
		`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
		`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email string `json:"email"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), sql.NullString{
		String: req.Email,
		Valid:  req.Email != "",
	})

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	RespondWithJSON(w, http.StatusOK, user)
}

func handlerValidation(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		Valid bool   `json:"valid"`
		Body  string `json:"body"`
		Error string `json:"error,omitempty"`
	}

	req := requestBody{}
	res := responseBody{
		Valid: true,
		Body:  req.Body,
		Error: "",
	}
	decoder := json.NewDecoder(r.Body)

	code := http.StatusOK

	err := decoder.Decode(&req)
	if err != nil {
		code = http.StatusBadRequest
		res.Valid = false
		res.Error = "Something went wrong"
	}

	if utf8.RuneCountInString(req.Body) > 140 {
		code = http.StatusBadRequest
		res.Valid = false
		res.Error = "Chirp is too long"
	}

	res.Body = cleanChirp(req.Body)

	dat, err := json.Marshal(res)
	if err != nil {
		code = http.StatusBadRequest
		res.Valid = false
		res.Error = "Something went wrong"
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(dat))
}

func cleanChirp(s string) string {
	s = strings.ToLower(s)

	s = strings.ReplaceAll(s, "kerfuffle", "****")
	s = strings.ReplaceAll(s, "sharbert", "****")
	s = strings.ReplaceAll(s, "fornax", "****")

	return s
}
