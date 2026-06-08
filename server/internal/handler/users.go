package handler

import (
	"context"
	"encoding/json"
	"errors"
	"lensamity/internal/core"
	"log/slog"
	"net/http"
	"time"
)

func (e *Env) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := core.GetUserProfile(ctx, e.Store, core.UserProfileDTO{Username: username})
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
