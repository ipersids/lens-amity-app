package handler

import (
	"context"
	"encoding/json"
	"errors"
	"lensamity/internal/users"
	"log/slog"
	"net/http"
	"time"
)

type UserHandler struct {
	userService *users.UserService
}

func NewUserHandler(userService *users.UserService) (*UserHandler, error) {
	if userService == nil {
		return nil, errors.New("handler: nil user service")
	}

	return &UserHandler{
		userService: userService,
	}, nil
}

type AvatarResponse struct {
	URL    string      `json:"url"`
	Method string      `json:"method"`
	Header http.Header `json:"header"`
}

type GetUserProfileResponse struct {
	Username    string          `json:"username"`
	DisplayName string          `json:"display_name"`
	PhotoCount  int64           `json:"photo_count"`
	Visibility  string          `json:"visibility"`
	JoinedAt    time.Time       `json:"joined_at"`
	Avatar      *AvatarResponse `json:"avatar,omitempty"`
	About       string          `json:"about,omitempty"`
}

func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.userService.GetUserProfile(ctx, username)
	if err != nil {
		if errors.Is(err, users.ErrorGetUserProfile) {
			slog.Error("UserProfile not found", "error", err)
			http.NotFound(w, r)
			return
		}
		slog.Error("UserProfile: request failed", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	profile := GetUserProfileResponse{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		About:       user.About,
		PhotoCount:  user.PhotoCount,
		Visibility:  user.Visibility,
		JoinedAt:    user.JoinedAt,
	}

	if user.AvatarPresignedRequest != nil {
		profile.Avatar = &AvatarResponse{
			Method: user.AvatarPresignedRequest.Method,
			URL:    user.AvatarPresignedRequest.URL,
			Header: user.AvatarPresignedRequest.SignedHeader,
		}
	}

	err = json.NewEncoder(w).Encode(profile)

	if err != nil {
		slog.Error("UserProfile handler: failed encode response", "error", err)
	}
}
