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

func (e *Env) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to decode SignupRequest", "error", err)
		http.Error(w, "malformed JSON", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "invalid credentials", http.StatusBadRequest)
		return
	}

	user, err := core.CreateUser(e.Store, core.CreateUserDTO{
		RawUsername:    req.Username,
		RawDisplayName: req.DisplayName,
		RawPassword:    req.Password,
	})

	if err != nil {
		if errors.Is(err, core.ErrorCreateUserFailed) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Error("Signup: request failed", "error", err)
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
		slog.Error("Signup: failed encode response", "error", err)
	}
}

func (e *Env) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	l := core.LoginDTO{RawUsername: req.Username, RawPassword: req.Password}

	data, err := core.Login(ctx, &e.Config.Auth, e.Store, l)
	if err != nil {
		slog.Error("Login request failed", "error", err)
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "Database timeout exceeded", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	err = json.NewEncoder(w).Encode(SuccessResponse{
		Data: LoginResponseBody{
			AccessToken:  data.Token,
			RefreshToken: data.RefreshToken,
			User: UserProfileReponseBody{
				Username:    data.User.UsernameKey,
				DisplayName: data.User.UsernameDisplay,
			},
		},
	})

	if err != nil {
		slog.Error("Login: failed encode response", "error", err)
	}
}
