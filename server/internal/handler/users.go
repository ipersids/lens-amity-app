package handler

import (
	"encoding/json"
	"errors"
	"lensamity/internal/core"
	"log/slog"
	"net/http"
)

func (e *Env) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	user, err := core.GetUserProfile(e.Store, core.UserProfileDTO{Username: username})
	if err != nil {
		if errors.Is(err, core.ErrorGetUserProfile) {
			slog.Error("UserProfile not found", "error", err)
			http.NotFound(w, r)
			return
		}
		slog.Error("UserProfile: request failed", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(SuccessResponse{
		Data: UserProfileReponseBody{
			Username:    user.UsernameKey,
			DisplayName: user.UsernameDisplay,
		},
	})

	if err != nil {
		slog.Error("UserProfile handler: failed encode response", "error", err)
	}
}
