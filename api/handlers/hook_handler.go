package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/portbound/bootdev-httpserver/api"
)

func UpgradeToChirpyRed(w http.ResponseWriter, r *http.Request, cfg *api.Config) {
	type hook struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
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
