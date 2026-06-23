package handler

import (
	"context"
	"encoding/json"
	"errors"
	"lensamity/internal/auth"
	"log/slog"
	"net/http"
	"time"
)

type UserHandler struct {
	userService *auth.UserService
}

func NewUserHandler(userService *auth.UserService) (*UserHandler, error) {
	if userService == nil {
		return nil, errors.New("handler: nil user service")
	}

	return &UserHandler{
		userService: userService,
	}, nil
}

type GetUserProfileResponse struct {
	Username    string
	DisplayName string
}

func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.userService.GetUserProfile(ctx, username)
	if err != nil {
		if errors.Is(err, auth.ErrorGetUserProfile) {
			slog.Error("UserProfile not found", "error", err)
			http.NotFound(w, r)
			return
		}
		slog.Error("UserProfile: request failed", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(GetUserProfileResponse{Username: user.UsernameKey, DisplayName: user.UsernameDisplay})

	if err != nil {
		slog.Error("UserProfile handler: failed encode response", "error", err)
	}
}
