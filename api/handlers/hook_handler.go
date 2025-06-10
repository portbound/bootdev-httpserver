package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/portbound/bootdev-httpserver/api"
	"github.com/portbound/bootdev-httpserver/internal/auth"
)

func UpgradeToChirpyRed(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	type hook struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		api.RespondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if key != cfg.PolkaKey {
		api.RespondWithError(w, http.StatusUnauthorized, fmt.Sprintln("Unauthorized"))
		return
	}

	h := hook{}
	if err := json.NewDecoder(r.Body).Decode(&h); err != nil {
		api.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if h.Event != "user.upgraded" {
		// quick exit if we're not upgrading th euser
		api.RespondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	if err := cfg.DbQueries.SetIsChirpyRed(r.Context(), h.Data.UserID); err != nil {
		api.RespondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	api.RespondWithJSON(w, http.StatusOK, nil)
}
