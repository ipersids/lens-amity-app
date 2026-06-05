package handler

import (
	"encoding/json"
	"errors"
	"lensamity/internal/core"
	"log/slog"
	"net/http"
)

func (e *Env) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("failed to decode SignupRequest", "error", err)
		// WriteError(w, http.StatusBadRequest, http.StatusText(http.StatusBadRequest), "Invalid email or password")
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

	err = json.NewEncoder(w).Encode(SignupReponse{
		Data: SignupReponseBody{
			Uuid:        user.Uuid.String(),
			Username:    user.UsernameKey,
			DisplayName: user.UsernameDisplay,
		},
	})

	if err != nil {
		slog.Error("Signup: failed encode response", "error", err)
	}
}
