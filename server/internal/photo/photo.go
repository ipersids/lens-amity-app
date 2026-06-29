package photo

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"lensamity/internal/storage"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type PhotoService struct {
	repo *photoRepository
	now  func() time.Time
}

func NewPhotoService(store *db.Store, s3Client *storage.Client) (*PhotoService, error) {
	if store == nil || s3Client == nil {
		return nil, errors.New("new photo service: nil postgres or s3 data")
	}

	repo, err := newPhotoRepository(store, s3Client.Bucket, s3Client.Presign)
	if err != nil {
		return nil, err
	}

	return &PhotoService{
		repo: repo,
		now:  time.Now,
	}, nil
}

type UploadObjectRequest struct {
	PhotoID uuid.UUID
	URL     string
	Method  string
	Header  map[string][]string
}

var (
	ErrInternal            = errors.New("internal error")
	ErrUnsupportedFileType = errors.New("unsupported file type")
	ErrPhotoAlreadyExists  = errors.New("photo already exists for date")
	ErrDateOutOfRange      = errors.New("date must be within the last 7 days including today")
)

type UploadObjectInput struct {
	UserID      uuid.UUID
	Date        time.Time
	ContentType string
}

func (ps *PhotoService) UploadObject(ctx context.Context, o UploadObjectInput) (UploadObjectRequest, error) {
	if ctx == nil {
		return UploadObjectRequest{}, fmt.Errorf("%w: nil context", ErrInternal)
	}
	if o.UserID == uuid.Nil {
		return UploadObjectRequest{}, errors.New("user id is required")
	}

	localDate := dateOnly(o.Date, time.UTC)
	if !withinLastSevenDays(localDate, ps.now().UTC()) {
		return UploadObjectRequest{}, ErrDateOutOfRange
	}

	objectID, err := uuid.NewV7()
	if err != nil {
		return UploadObjectRequest{}, fmt.Errorf("%w: generate upload object uuid failed: %w", ErrInternal, err)
	}

	ext, err := imageExt(o.ContentType)
	if err != nil {
		return UploadObjectRequest{}, err
	}

	objectKey := fmt.Sprintf(
		"photos/%s/%s/%s/original.%s",
		o.UserID.String(),
		localDate.Format("2006-01-02"),
		objectID.String(),
		ext,
	)
	expiresAt := ps.now().UTC().Add(10 * time.Minute)

	res, err := ps.repo.createPendingUpload(ctx, createPendingUploadParams{
		PhotoID:     objectID,
		OwnerUserID: o.UserID,
		ObjectKey:   objectKey,
		LocalDate:   localDate,
		ContentType: o.ContentType,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return UploadObjectRequest{}, ErrPhotoAlreadyExists
		}
		return UploadObjectRequest{}, fmt.Errorf("%w: create upload intent: %w", ErrInternal, err)
	}

	return *res, nil
}

func imageExt(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	case "image/webp":
		return "webp", nil
	default:
		return "", ErrUnsupportedFileType
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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
