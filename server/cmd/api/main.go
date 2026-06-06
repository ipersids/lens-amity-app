package main

import (
	"context"
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

func registerRoutes(mux *http.ServeMux, app *handler.Env) {
	mux.HandleFunc("GET /health", app.HealthCheck)

	mux.HandleFunc("POST /api/signup", app.Signup)
	mux.HandleFunc("GET /api/users/{id}", app.UserProfile)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	store, err := db.InitStore(ctx)
	if err != nil {
		slog.Error("store initialisation failed", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	app := &handler.Env{
		Store: store,
	}

	mux := http.NewServeMux()
	registerRoutes(mux, app)

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
