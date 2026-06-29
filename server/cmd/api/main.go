package main

import (
	"context"
	"lensamity/internal/auth"
	"lensamity/internal/config"
	"lensamity/internal/db"
	"lensamity/internal/handler"
	"lensamity/internal/middleware"
	"lensamity/internal/photo"
	"lensamity/internal/storage"
	"lensamity/internal/users"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type handlers struct {
	auth  *handler.AuthHandler
	user  *handler.UserHandler
	photo *handler.PhotoHandler
}

func (h *handlers) registerRoutes(mux *http.ServeMux, authRequiredMiddleware func(http.HandlerFunc) http.HandlerFunc) {
	// 1. Public routes
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("POST /api/auth/signup", h.auth.Signup)
	mux.HandleFunc("POST /api/auth/login", h.auth.Login)
	mux.HandleFunc("POST /api/auth/logout", h.auth.Logout)
	mux.HandleFunc("POST /api/auth/logout-all", authRequiredMiddleware(h.auth.LogoutAll))

	// 2. @TODO Context-aware Profile route
	mux.HandleFunc("GET /api/users/{username}", authRequiredMiddleware(h.user.GetUserProfile))

	// 3. Strict Session Protected routes
	mux.HandleFunc("PUT /api/photos/upload", authRequiredMiddleware(h.photo.UploadIntent))
	// mux.HandleFunc("PUT /api/photos/upload/complete", authRequiredMiddleware(h.photo.UploadComplete))

	// mux.Handle("GET /api/users/me", middleware.RequireAuth(http.HandlerFunc(app.GetMyProfile)))
	// mux.Handle("PATCH /api/users/me", middleware.RequireAuth(http.HandlerFunc(app.UpdateMyProfile)))
	// mux.Handle("DELETE /api/users/me", middleware.RequireAuth(http.HandlerFunc(app.DeleteMyProfile)))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	conf, err := config.Load()
	if err != nil {
		slog.Error("Loading envirenment failed", "error", err)
		os.Exit(1)
	}

	store, err := db.NewStore(ctx, conf.DatabaseURL)
	if err != nil {
		slog.Error("DB store initialisation failed", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	storage, err := storage.NewS3Client(conf.S3)
	if err != nil {
		slog.Error("S3 storage initialisation failed", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	authService, err := auth.NewAuthService(store, conf.SessionSecret)
	if err != nil {
		slog.Error("auth service initialisation failed", "error", err)
		os.Exit(1)
	}

	userService, err := users.NewUserService(store)
	if err != nil {
		slog.Error("user service initialisation failed", "error", err)
		os.Exit(1)
	}

	photoService, err := photo.NewPhotoService(store, storage)
	if err != nil {
		slog.Error("photo service initialisation failed", "error", err)
		os.Exit(1)
	}

	authHandler, err := handler.NewAuthHandler(authService)
	if err != nil {
		slog.Error("auth handler initialisation failed", "error", err)
		os.Exit(1)
	}

	userHandler, err := handler.NewUserHandler(userService)
	if err != nil {
		slog.Error("user handler initialisation failed", "error", err)
		os.Exit(1)
	}

	photoHandler, err := handler.NewPhotoHandler(photoService)
	if err != nil {
		slog.Error("photo handler initialisation failed", "error", err)
		os.Exit(1)
	}

	h := handlers{
		auth:  authHandler,
		user:  userHandler,
		photo: photoHandler,
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
