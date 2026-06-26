package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"lensamity/internal/middleware"
	"lensamity/internal/photo"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type photoService interface {
	UploadObject(context.Context, photo.UploadObjectInput) (photo.UploadObjectRequest, error)
}

type PhotoHandler struct {
	photoService photoService
}

func NewPhotoHandler(service photoService) (*PhotoHandler, error) {
	if service == nil {
		return nil, errors.New("photo handler: nil photo service")
	}
	v := reflect.ValueOf(service)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, errors.New("photo handler: nil photo service")
	}

	return &PhotoHandler{
		photoService: service,
	}, nil
}

type UploadIntentRequest struct {
	Date        string `json:"date"`
	ContentType string `json:"content_type"`
}

type UploadResponseBody struct {
	URL     string              `json:"url"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
}

func (ph *PhotoHandler) UploadIntent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)
	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), "Unauthorized")
		return
	}

	var req UploadIntentRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "malformed_json", "malformed JSON")
		return
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		WriteError(w, http.StatusBadRequest, "malformed_json", "malformed JSON")
		return
	}

	if req.Date == "" || req.ContentType == "" {
		WriteError(w, http.StatusBadRequest, "invalid_upload_intent", "date and content_type are required")
		return
	}

	date, err := time.Parse("02-01-2006", req.Date)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_date", "date must be DD-MM-YYYY")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := ph.photoService.UploadObject(ctx, photo.UploadObjectInput{
		UserID:      userID,
		Date:        date,
		ContentType: req.ContentType,
	})
	if err != nil {
		if errors.Is(err, photo.ErrUnsupportedFileType) {
			WriteError(w, http.StatusBadRequest, "unsupported_file_type", "unsupported file type")
			return
		}
		if errors.Is(err, photo.ErrDateOutOfRange) {
			WriteError(w, http.StatusBadRequest, "date_out_of_range", "date must be within the last 7 days including today")
			return
		}
		if errors.Is(err, photo.ErrPhotoAlreadyExists) {
			WriteError(w, http.StatusConflict, "photo_already_exists", "photo already exists for date")
			return
		}
		slog.Error("UploadIntent: request failed", "error", err)
		WriteError(w, statusForPhotoError(err), "internal_error", "something went wrong")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(UploadResponseBody{
		URL:     res.URL,
		Method:  res.Method,
		Headers: res.Header,
	})

	if err != nil {
		slog.Error("UploadIntent: failed encode response", "error", err)
	}
}

func statusForPhotoError(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}
