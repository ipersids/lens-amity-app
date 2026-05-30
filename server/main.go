package main

import (
	"errors"
	"fmt"
	"lensamity/internal/handler"
	"lensamity/internal/middleware"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", handler.HealthCheck)
}

func main() {
	mux := http.NewServeMux()
	registerRoutes(mux)

	handler := middleware.Logging(mux)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("server starting", "port", server.Addr)
	err := server.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server closed")
	} else if err != nil {
		fmt.Printf("Error starting server: %s\n", err)
		os.Exit(1)
	}
}
