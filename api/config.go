package api

import (
	"database/sql"
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
	PolkaKey       string
}

func NewConfig() (*Config, error) {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	return &Config{
		JWT:       os.Getenv("JWT"),
		PolkaKey:  os.Getenv("POLKA_KEY"),
		DbQueries: database.New(db)}, nil
}

func (cfg *Config) HandlerReadiness(w http.ResponseWriter, r *http.Request) {
	RespondWithJSON(w, http.StatusOK, "OK")
}

func (cfg *Config) HandlerReset(w http.ResponseWriter, r *http.Request) {
	// TODO: reset all tables?
}
