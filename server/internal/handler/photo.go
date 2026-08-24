package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"lensamity/internal/middleware"
	"lensamity/internal/uploads"
	"log/slog"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type photoService interface {
	UploadPhotoIntent(ctx context.Context, p uploads.UploadPhotoIntentParams) (*uploads.UploadPhotoIntentResult, error)
	UploadPhotoComplete(ctx context.Context, p uploads.UploadPhotoCompleteParams) error
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
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UploadResponseBody struct {
	PhotoID uuid.UUID   `json:"photoID"`
	URL     string      `json:"url"`
	Method  string      `json:"method"`
	Header  http.Header `json:"header"`
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

	if req.Size <= 0 {
		WriteError(w, http.StatusBadRequest, "invalid_upload_intent", "size is negative or zero")
		return
	}

	date, err := time.Parse("02-01-2006", req.Date)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid_date", "date must be DD-MM-YYYY")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	res, err := ph.photoService.UploadPhotoIntent(ctx, uploads.UploadPhotoIntentParams{
		OwnerUserID: userID,
		PhotoDate:   date,
		ContentType: req.ContentType,
		Size:        req.Size,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, uploads.ErrUnsupportedFileType) {
			WriteError(w, http.StatusBadRequest, "unsupported_file_type", "unsupported file type")
			return
		}
		if errors.Is(err, uploads.ErrDateOutOfRange) {
			WriteError(w, http.StatusBadRequest, "date_out_of_range", "date must be within the last 7 days including today")
			return
		}
		if errors.Is(err, uploads.ErrPhotoAlreadyExists) {
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
		URL:     res.PresignedRequest.URL,
		Method:  res.PresignedRequest.Method,
		Header:  res.PresignedRequest.SignedHeader,
		PhotoID: res.PhotoID,
	})

	if err != nil {
		slog.Error("UploadIntent: failed encode response", "error", err)
	}
}

func (ph *PhotoHandler) UploadComplete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	photoID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "malformed_id", "malformed ID")
		return
	}

	ctx := r.Context()

	userID, ok := ctx.Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		WriteError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized), "Unauthorized")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := ph.photoService.UploadPhotoComplete(ctx, uploads.UploadPhotoCompleteParams{ID: photoID, OwnerUserID: userID}); err != nil {
		switch {
		case errors.Is(err, uploads.ErrUnsupportedFileType):
			WriteError(w, http.StatusUnsupportedMediaType, "unsupported_file_type", "unsupported file type")
		case errors.Is(err, uploads.ErrFileTooLarge):
			WriteError(w, http.StatusRequestEntityTooLarge, "file_too_large", "file must be 10 MB or smaller")
		case errors.Is(err, uploads.ErrPhotoNotFound):
			WriteError(w, http.StatusNotFound, "photo_not_found", "photo not found")
		case errors.Is(err, uploads.ErrPhotoAlreadyExists):
			WriteError(w, http.StatusConflict, "photo_not_completable", "photo cannot be completed")
		case errors.Is(err, uploads.ErrUploadNotFound):
			WriteError(w, http.StatusNotFound, "upload_not_found", "uploaded object not found")
		default:
			slog.Error("UploadComplete: request failed", "error", err)
			WriteError(w, statusForPhotoError(err), "internal_error", "something went wrong")
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func statusForPhotoError(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusInternalServerError
}
