package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/portbound/bootdev-httpserver/internal/database"
)

type Config struct {
	FileserverHits atomic.Int32
	DbQueries      *database.Queries
	JWT            string
}

func NewConfig() (*Config, error) {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	return &Config{DbQueries: database.New(db)}, nil
}

func (cfg *Config) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *Config) HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, "text/plain; charset=utf-8", "OK")
}

func (cfg *Config) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	msg := fmt.Sprintf(
		`<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
		`, cfg.FileserverHits.Load())

	RespondWithJSON(w, http.StatusOK, "text/html; charset=utf-8", msg)
}

func (cfg *Config) HandlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.FileserverHits.Store(0)
	RespondWithJSON(w, http.StatusOK, "text/plain; charset=utf-8", "Hits reset to 0")
}
