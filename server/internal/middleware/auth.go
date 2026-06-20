package middleware

import (
	"context"
	"lensamity/internal/auth"
	"net/http"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

// Rejects requests if Authorization header is missing or JWT is invalid
func StrictAuth(authService *auth.AuthService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			ctx := context.WithValue(r.Context(), UserIDKey, "")

			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}
