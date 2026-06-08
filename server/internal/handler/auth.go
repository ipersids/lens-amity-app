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

type AuthHandler struct {
	authService *core.AuthService
}

func NewAuthHandler(authService *core.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

type SignupRequest struct {
	Username    string `json:"username" validate:"required"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password" validate:"required"`
}

type SignupResponse struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		slog.Error("failed to decode SignupRequest", "error", err)
		http.Error(w, "malformed JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "invalid credentials", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.authService.Signup(ctx, req.Username, req.DisplayName, req.Password)

	if err != nil {
		if errors.Is(err, core.ErrorCreateUserFailed) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("Signup: request failed", "error", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(SignupResponse{Username: user.UsernameKey, DisplayName: user.UsernameDisplay})

	if err != nil {
		slog.Error("Signup: failed encode response", "error", err)
	}
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		slog.Error("failed to decode LoginRequest", "error", err)
		http.Error(w, "malformed JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "invalid credentials", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token, refreshToken, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		slog.Error("Login request failed", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Database timeout exceeded", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	err = json.NewEncoder(w).Encode(LoginResponse{AccessToken: token, RefreshToken: refreshToken})

	if err != nil {
		slog.Error("Login: failed encode response", "error", err)
	}
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		slog.Error("failed to decode RefreshRequest", "error", err)
		http.Error(w, "malformed request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "malformed request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	accessToken, refreshToken, err := h.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Database timeout exceeded", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	err = json.NewEncoder(w).Encode(RefreshResponse{AccessToken: accessToken, RefreshToken: refreshToken})

	if err != nil {
		slog.Error("Refresh: failed encode response", "error", err)
	}
}
