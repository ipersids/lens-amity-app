package handler

import (
	"context"
	"encoding/json"
	"errors"
	"lensamity/internal/auth"
	"lensamity/internal/middleware"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

const maxAuthBodyBytes = 8 * 1024 // 8 KiB

type SignupRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
}

type SignupResponse struct {
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)

	var req SignupRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "malformed_json", "malformed JSON")
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "invalid_signup", "username and password are required")
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.authService.Signup(ctx, req.Username, req.DisplayName, req.Password)

	if err != nil {
		if errors.Is(err, auth.ErrUsernameTaken) {
			WriteError(w, http.StatusConflict, "username_taken", "username is not available")
			return
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, auth.ErrInternal) {
			slog.Error("Signup: request failed", "error", err)
			WriteError(w, statusForAuthError(err), "internal_error", "something went wrong")
			return
		}
		WriteError(w, http.StatusBadRequest, "invalid_signup", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(SignupResponse{Username: user.UsernameKey, DisplayName: user.UsernameDisplay})

	if err != nil {
		slog.Error("Signup: failed encode response", "error", err)
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Username     string `json:"username"`
	DisplayName  string `json:"displayName"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)

	var req LoginRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "malformed_json", "malformed JSON")
		return
	}

	if req.Username == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "username and password are required")
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			WriteError(w, http.StatusUnauthorized, "invalid_credentials", "")
			return
		}
		slog.Error("Login request failed", "error", err)
		WriteError(w, statusForAuthError(err), "internal_error", "something went wrong")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(LoginResponse{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
	})

	if err != nil {
		slog.Error("Login: failed encode response", "error", err)
	}
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)

	var req RefreshRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "malformed_json", "malformed request body")
		return
	}

	if req.RefreshToken == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "refresh token is required")
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	refreshed, err := h.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) || errors.Is(err, auth.ErrCompromisedToken) {
			if errors.Is(err, auth.ErrCompromisedToken) {
				slog.Warn("refresh token replay detected", "error", err)
			}
			WriteError(w, http.StatusUnauthorized, "invalid_token", "")
			return
		}
		slog.Error("Refresh request failed", "error", err)
		WriteError(w, statusForAuthError(err), "internal_error", "something went wrong")
		return
	}

	if refreshed.Replayed {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(RefreshResponse{AccessToken: refreshed.AccessToken, RefreshToken: refreshed.RefreshToken})

	if err != nil {
		slog.Error("Refresh: failed encode response", "error", err)
	}
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)

	var req RefreshRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "malformed_json", "malformed request body")
		return
	}

	if req.RefreshToken == "" {
		WriteError(w, http.StatusBadRequest, "invalid_request", "refresh token is required")
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := h.authService.Logout(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidToken) || errors.Is(err, auth.ErrExpiredToken) {
			WriteError(w, http.StatusUnauthorized, "invalid_token", "")
			return
		}
		slog.Error("Logout request failed", "error", err)
		WriteError(w, statusForAuthError(err), "internal_error", "something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)

	userID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := h.authService.LogoutAll(ctx, userID)
	if err != nil {
		slog.Error("LogoutAll request failed", "error", err)
		WriteError(w, statusForAuthError(err), "internal_error", "something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func statusForAuthError(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}
