package middleware

import (
	"log/slog"
	"net/http"
)

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("RequireAuth MIDDLEWARE")
		next.ServeHTTP(w, r)
	})
}
