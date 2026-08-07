package uploads

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"lensamity/internal/storage"
	"mime"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/google/uuid"
)

var (
	ErrInternal            = errors.New("internal error")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrFileTooLarge        = errors.New("file size exceeded limits")
	ErrPhotoNotFound       = errors.New("photo not found")
	ErrPhotoNotCompletable = errors.New("photo cannot be completed")
	ErrUploadNotFound      = errors.New("uploaded object not found")
	ErrPhotoAlreadyExists  = errors.New("photo already exists for date")
	ErrDateOutOfRange      = errors.New("date must be within the last 7 days including today")
)

const maxImageBytes int64 = 10 * 1024 * 1024

type PhotoService struct {
	repo photoStore
	now  func() time.Time
}

type photoStore interface {
	presignPutObject(ctx context.Context, p presignPutObjectParams) (*v4.PresignedHTTPRequest, error)
	createPendingPhotoUpload(ctx context.Context, p createPendingPhotoUploadParams) error
	pendingPhotoUpload(ctx context.Context, p pendingPhotoUploadParams) (*pendingPhotoUploadData, error)
	completePendingPhotoUpload(ctx context.Context, p completePendingPhotoUploadParams) error
	failPendingPhotoUpload(ctx context.Context, p failPendingPhotoUploadParams) error
}

func NewPhotoService(store *db.Store, s3Client *storage.Client) (*PhotoService, error) {
	if store == nil || s3Client == nil {
		return nil, errors.New("new photo service: nil postgres or s3 data")
	}

	repo, err := newUploadRepository(store, s3Client)
	if err != nil {
		return nil, err
	}

	return &PhotoService{
		repo: repo,
		now:  time.Now,
	}, nil
}

type UploadPhotoIntentParams struct {
	OwnerUserID uuid.UUID
	PhotoDate   time.Time
	ContentType string
	Size        int64
	Title       string
	Description string
}

func (ps *PhotoService) UploadPhotoIntent(ctx context.Context, p UploadPhotoIntentParams) (*v4.PresignedHTTPRequest, error) {
	if ctx == nil {
		return nil, fmt.Errorf("%w: nil context", ErrInternal)
	}

	validated, err := ps.validateUploadPhotoIntentParams(&p)
	if err != nil {
		return nil, err
	}

	photoID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%w: generate photo upload uuid: %w", ErrInternal, err)
	}

	key := fmt.Sprintf(
		"%s/photos/%s/%s/original.%s",
		validated.OwnerUserID.String(),
		validated.PhotoDate.Format("2006-01-02"),
		photoID.String(),
		validated.Extension,
	)
	expiresAt := ps.now().UTC().Add(10 * time.Minute)

	req, err := ps.repo.presignPutObject(ctx, presignPutObjectParams{
		Key:         key,
		ContentType: validated.ContentType,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: create presigned photo upload: %w", ErrInternal, err)
	}

	err = ps.repo.createPendingPhotoUpload(ctx, createPendingPhotoUploadParams{
		ID:          photoID,
		OwnerUserID: validated.OwnerUserID,
		Key:         key,
		ContentType: validated.ContentType,
		Size:        validated.Size,
		PhotoDate:   validated.PhotoDate,
		Title:       validated.Title,
		Description: validated.Description,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: create pending photo upload: %w", ErrInternal, err)
	}

	return req, nil
}

type validatedUploadPhotoIntentParams struct {
	OwnerUserID uuid.UUID
	PhotoDate   time.Time
	ContentType string
	Extension   string
	Size        int64
	Title       string
	Description string
}

func (ps *PhotoService) validateUploadPhotoIntentParams(p *UploadPhotoIntentParams) (validatedUploadPhotoIntentParams, error) {
	if ps == nil {
		return validatedUploadPhotoIntentParams{}, fmt.Errorf("%w: nil photo service", ErrInternal)
	}
	if p == nil {
		return validatedUploadPhotoIntentParams{}, fmt.Errorf("%w: nil upload photo intent params", ErrInternal)
	}
	if p.OwnerUserID == uuid.Nil {
		return validatedUploadPhotoIntentParams{}, fmt.Errorf("%w: owner user id is required", ErrInternal)
	}

	photoDate := dateOnly(p.PhotoDate, time.UTC)
	if !withinLastSevenDays(photoDate, ps.now().UTC()) {
		return validatedUploadPhotoIntentParams{}, ErrDateOutOfRange
	}

	contentType, ext, err := normalizeImageContentType(p.ContentType)
	if err != nil {
		return validatedUploadPhotoIntentParams{}, err
	}

	if p.Size <= 0 || p.Size > maxImageBytes {
		return validatedUploadPhotoIntentParams{}, ErrFileTooLarge
	}

	return validatedUploadPhotoIntentParams{
		OwnerUserID: p.OwnerUserID,
		PhotoDate:   photoDate,
		ContentType: contentType,
		Extension:   ext,
		Size:        p.Size,
		Title:       p.Title,
		Description: p.Description,
	}, nil
}

func normalizeImageContentType(contentType string) (string, string, error) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", "", ErrUnsupportedFileType
	}

	switch mediaType {
	case "image/jpeg":
		return mediaType, "jpg", nil
	case "image/png":
		return mediaType, "png", nil
	case "image/webp":
		return mediaType, "webp", nil
	case "image/heic":
		return mediaType, "heic", nil
	case "image/heif":
		return mediaType, "heif", nil
	default:
		return "", "", ErrUnsupportedFileType
	}
}

func dateOnly(t time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	t = t.In(loc)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func withinLastSevenDays(date time.Time, now time.Time) bool {
	today := dateOnly(now, date.Location())
	minDate := today.AddDate(0, 0, -6)
	return !date.Before(minDate) && !date.After(today)
}
