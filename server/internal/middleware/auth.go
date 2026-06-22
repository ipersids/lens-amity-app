package middleware

import (
	"context"
	"errors"
	"lensamity/internal/auth"
	"log/slog"
	"net/http"
	"time"
)

type contextKey string

const (
	UserIDKey         contextKey = "userID"
	SessionCookieName            = "session"
)

func SetSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

func StrictAuth(authService *auth.AuthService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				ClearSessionCookie(w)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			session, err := authService.ValidateSession(ctx, cookie.Value)
			if err != nil {
				if errors.Is(err, auth.ErrInvalidSession) {
					ClearSessionCookie(w)
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				slog.Error("session validation failed", "error", err)
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			requestCtx := context.WithValue(r.Context(), UserIDKey, session.UserID)
			next.ServeHTTP(w, r.WithContext(requestCtx))
		}
	}
}
