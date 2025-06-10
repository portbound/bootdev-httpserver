package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/portbound/bootdev-httpserver/api"
	"github.com/portbound/bootdev-httpserver/internal/auth"
	"github.com/portbound/bootdev-httpserver/internal/database"
)

func RefreshAccessToken(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	type response struct {
		Token string `json:"token"`
	}

	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	refTok, err := cfg.DbQueries.GetRefreshToken(r.Context(), tok)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if refTok.RevokedAt.Valid {
		api.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Refresh token has been revoked"))
		return
	}

	jwt, err := auth.MakeJWT(refTok.UserID, cfg.JWT)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", err))
		return
	}

	resp := response{
		Token: jwt,
	}
	api.RespondWithJSON(w, http.StatusOK, resp)
}

func RevokeRefreshToken(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
	}

	if err := cfg.DbQueries.RevokeRefreshToken(r.Context(), tok); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
	}
	api.RespondWithJSON(w, http.StatusNoContent, nil)
}

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
		Token          string    `json:"token"`
		RefreshToken   string    `json:"refresh_token"`
		IsChirpyRed    bool      `json:"is_chirpy_red"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := cfg.DbQueries.GetUser(r.Context(), req.Email)
	if err != nil {
		api.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Could not find user with email: %s", req.Email))
		return
	}

	if err := auth.CheckPasswordHash(user.HashedPassword, req.Password); err != nil {
		api.RespondWithError(w, http.StatusNotFound, fmt.Sprintf("Invalid password for: %s", req.Email))
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.JWT)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", err))
		return
	}

	tok := auth.MakeRefreshToken()

	params := database.CreateRefreshTokenParams{
		Token:  tok,
		UserID: user.ID,
	}
	refreshToken, err := cfg.DbQueries.CreateRefreshToken(r.Context(), params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Something went wrong: %s", err))
	}

	resp := response{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt.Time,
		UpdatedAt:    user.UpdatedAt.Time,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: refreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	}
	api.RespondWithJSON(w, http.StatusOK, resp)
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
		IsChirpyRed    bool      `json:"is_chirpy_red"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hashedPasswd, err := auth.HashPassword(req.Password)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "failed to set password")
		return
	}

	params := database.CreateUserParams{
		Email:          req.Email,
		HashedPassword: hashedPasswd,
	}

	user, err := cfg.DbQueries.CreateUser(r.Context(), params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	resp := response{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt.Time,
		UpdatedAt:   user.UpdatedAt.Time,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
	api.RespondWithJSON(w, http.StatusCreated, resp)
}

func UpdateUser(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := database.UpdateUserParams{}
	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)

	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	params.ID, err = auth.ValidateJWT(tok, cfg.JWT)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, err.Error())
	}

	if req.Password != "" {
		params.HashedPassword, err = auth.HashPassword(req.Password)
		if err != nil {
			api.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}

	if req.Email != "" {
		params.Email = req.Email
	}

	updatedUser, err := cfg.DbQueries.UpdateUser(r.Context(), params)
	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	api.RespondWithJSON(w, http.StatusOK, updatedUser)
}
