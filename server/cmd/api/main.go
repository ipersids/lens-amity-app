package main

import (
	"context"
	"lensamity/internal/core"
	"lensamity/internal/db"
	"lensamity/internal/handler"
	"lensamity/internal/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type handlers struct {
	auth *handler.AuthHandler
	user *handler.UserHandler
}

func (h *handlers) registerRoutes(mux *http.ServeMux, authRequiredMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// 1. Public routes
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("POST /api/auth/signup", h.auth.Signup)
	mux.HandleFunc("POST /api/auth/login", h.auth.Login)
	mux.HandleFunc("POST /api/auth/refresh", h.auth.Refresh)
	mux.HandleFunc("POST /api/auth/logout", authRequiredMiddleware(h.auth.Logout))
	mux.HandleFunc("POST /api/auth/logout-all", authRequiredMiddleware(h.auth.LogoutAll))

	// 2. @TODO Context-aware Profile route
	mux.HandleFunc("GET /api/users/{username}", authRequiredMiddleware(h.user.GetUserProfile))

	// 3. Strict Session Protected routes
	// mux.Handle("GET /api/users/me", middleware.RequireAuth(http.HandlerFunc(app.GetMyProfile)))
	// mux.Handle("PATCH /api/users/me", middleware.RequireAuth(http.HandlerFunc(app.UpdateMyProfile)))
	// mux.Handle("DELETE /api/users/me", middleware.RequireAuth(http.HandlerFunc(app.DeleteMyProfile)))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf, err := core.InitConfig()
	if err != nil {
		slog.Error("config initialisation failed", "error", err)
		os.Exit(1)
	}

	store, err := db.InitStore(ctx, conf.DatabaseURL)
	if err != nil {
		slog.Error("store initialisation failed", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	mux := http.NewServeMux()

	authService := core.NewAuthService(store, &conf.Auth)
	userService := core.NewUserService(store)

	h := handlers{
		auth: handler.NewAuthHandler(authService),
		user: handler.NewUserHandler(userService),
	}

	h.registerRoutes(mux, middleware.StrictAuth(authService))

	handler := middleware.Logging(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}
}
