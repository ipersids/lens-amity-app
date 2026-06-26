package handler

import (
	"context"
	"encoding/json"
	"errors"
	"lensamity/internal/auth"
	"lensamity/internal/middleware"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type authService interface {
	Signup(context.Context, string, string, string) (*auth.SignupResponse, error)
	Login(context.Context, string, string) (*auth.LoginResult, error)
	Logout(context.Context, string) error
	LogoutAll(context.Context, uuid.UUID) error
}

type AuthHandler struct {
	authService authService
}

func NewAuthHandler(service authService) (*AuthHandler, error) {
	if service == nil {
		return nil, errors.New("handler: nil auth service")
	}
	v := reflect.ValueOf(service)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, errors.New("handler: nil auth service")
	}

	return &AuthHandler{
		authService: service,
	}, nil
}

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
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
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

	middleware.SetSessionCookie(w, user.CookieToken, user.CookieExpiredAt)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(LoginResponse{
		Username:    user.Username,
		DisplayName: user.DisplayName,
	})

	if err != nil {
		slog.Error("Login: failed encode response", "error", err)
	}
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cookie, err := r.Cookie(middleware.SessionCookieName)
	if errors.Is(err, http.ErrNoCookie) {
		middleware.ClearSessionCookie(w)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		middleware.ClearSessionCookie(w)
		WriteError(w, http.StatusBadRequest, "invalid_cookie", "invalid session cookie")
		return
	}

	err = h.authService.Logout(ctx, cookie.Value)
	if err != nil {
		slog.Error("Logout request failed", "error", err)
		WriteError(w, statusForAuthError(err), "internal_error", "something went wrong")
		return
	}

	middleware.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
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

	middleware.ClearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func statusForAuthError(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}
