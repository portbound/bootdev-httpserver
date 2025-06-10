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
	"github.com/portbound/bootdev-httpserver/internal/auth"
	"github.com/portbound/bootdev-httpserver/internal/database"
)

type Chirp struct {
	Body string `json:"body"`
}

func CreateChirp(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	validUserID, err := auth.ValidateJWT(tok, cfg.JWT)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Unable to validate token: %s", err))
		return
	}

	chirp := &Chirp{}
	if err := json.NewDecoder(r.Body).Decode(chirp); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = chirp.validate()
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	chirp.clean()

	params := database.CreateChirpParams{
		Body:   sql.NullString{String: chirp.Body, Valid: chirp.Body != ""},
		UserID: validUserID,
	}

	createdChirp, err := cfg.DbQueries.CreateChirp(r.Context(), params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create Chirp: %s", err))
		return
	}

	api.RespondWithJSON(w, http.StatusCreated, "application/json", createdChirp)
}

func GetAllChirps(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	chirps, err := cfg.DbQueries.GetAllChirps(r.Context())
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch chirps: %s", err))
		return
	}
	api.RespondWithJSON(w, http.StatusOK, "application/json", chirps)
}

func GetChirp(w http.ResponseWriter, r *http.Request, cfg *api.Config, chirpID uuid.UUID) {
	chirp, err := cfg.DbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		api.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Chirp not found: %s", err))
		return
	}
	api.RespondWithJSON(w, http.StatusOK, "application/json", chirp)
}

func DeleteChirp(w http.ResponseWriter, r *http.Request, cfg *api.Config, chirpID uuid.UUID) {
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	userID, err := auth.ValidateJWT(tok, cfg.JWT)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	chirp, err := cfg.DbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		api.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	if chirp.UserID != userID {
		api.RespondWithError(w, http.StatusForbidden, fmt.Sprint("Unauthorized. User does not own chirp"))
		return
	}

	if err := cfg.DbQueries.DeleteChirp(r.Context(), chirpID); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
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
