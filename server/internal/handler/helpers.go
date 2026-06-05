package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	})

	if err != nil {
		slog.Error("failed WriteError", "error", err)
	}
}
