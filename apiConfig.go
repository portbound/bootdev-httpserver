package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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
