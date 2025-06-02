package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/portbound/bootdev-httpserver/api"
	"github.com/portbound/bootdev-httpserver/internal/auth"
	"github.com/portbound/bootdev-httpserver/internal/database"
)

func Login(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	type request struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		ID             uuid.UUID `json:"id"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		Email          string    `json:"email"`
		HashedPassword string    `json:"-"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := cfg.DbQueries.GetUser(r.Context(), sql.NullString{
		String: req.Email,
		Valid:  req.Email != "",
	})
	if err != nil {
		api.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Could not find user with email: %s", req.Email))
		return
	}

	if err := auth.CheckPasswordHash(user.HashedPassword.String, req.Password); err != nil {
		api.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Invalid password for: %s", req.Email))
		return
	}

	resp := response{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email.String,
	}
	api.RespondWithJSON(w, http.StatusOK, "application/json", resp)
}

func CreateUser(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	type request struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	type response struct {
		ID             uuid.UUID `json:"id"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		Email          string    `json:"email"`
		HashedPassword string    `json:"-"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "failed to set password")
		return
	}

	params := database.CreateUserParams{
		Email: sql.NullString{
			String: req.Email,
			Valid:  req.Email != "",
		},
		HashedPassword: sql.NullString{
			String: hash,
			Valid:  hash != "",
		},
	}

	user, err := cfg.DbQueries.CreateUser(r.Context(), params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	resp := response{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Email:     user.Email.String,
	}
	api.RespondWithJSON(w, http.StatusCreated, "application/json", resp)
}
