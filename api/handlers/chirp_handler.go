package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/portbound/bootdev-httpserver/api"
	"github.com/portbound/bootdev-httpserver/internal/database"
)

type Chirp struct {
	Body   string `json:"body"`
	UserId string `json:"user_id"`
}

func CreateChirp(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	chirp := &Chirp{}
	if err := json.NewDecoder(r.Body).Decode(chirp); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := chirp.validate()
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	chirp.clean()

	params := database.CreateChirpParams{
		Body:   sql.NullString{String: chirp.Body, Valid: chirp.Body != ""},
		UserID: uuid.MustParse(chirp.UserId),
	}

	createdChirp, err := cfg.DbQueries.CreateChirp(r.Context(), params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create Chirp: %s", err))
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "application/json", createdChirp)
}

func (c *Chirp) validate() error {
	if utf8.RuneCountInString(c.Body) > 140 {
		return errors.New("Invalid Chirp Length")
	}
	return nil
}

func (c *Chirp) clean() {
	s := strings.ToLower(c.Body)
	s = strings.ReplaceAll(s, "kerfuffle", "****")
	s = strings.ReplaceAll(s, "sharbert", "****")
	s = strings.ReplaceAll(s, "fornax", "****")
	c.Body = s
}
