package handler

import (
	"context"
	"encoding/json"
	"errors"
	"lensamity/internal/middleware"
	"lensamity/internal/users"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	Username      string          `json:"username"`
	DisplayName   string          `json:"displayName"`
	PhotoCount    int64           `json:"photoCount"`
	CanEdit       bool            `json:"canEdit"`
	CanViewPhotos bool            `json:"canViewPhotos"`
	JoinedAt      time.Time       `json:"joinedAt"`
	Avatar        *AvatarResponse `json:"avatar,omitempty"`
	About         string          `json:"about,omitempty"`
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

	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	profile := GetUserProfileResponse{
		Username:      user.Username,
		DisplayName:   user.DisplayName,
		About:         user.About,
		PhotoCount:    user.PhotoCount,
		CanEdit:       user.ID == userID,
		CanViewPhotos: user.Visibility == "public",
		JoinedAt:      user.JoinedAt,
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
