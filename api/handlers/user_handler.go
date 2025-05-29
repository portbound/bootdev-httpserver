package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/portbound/bootdev-httpserver/api"
)

func CreateUser(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	type request struct {
		Email string `json:"email"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := cfg.DbQueries.CreateUser(r.Context(), sql.NullString{
		String: req.Email,
		Valid:  req.Email != "",
	})

	if err != nil {
		api.RespondWithError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	api.RespondWithJSON(w, http.StatusOK, "application/json", user)
}
