package uploads

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"lensamity/internal/storage"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type uploadRepository struct {
	store *db.Store
	s3    *storage.Client
}

func newUploadRepository(store *db.Store, s3 *storage.Client) (*uploadRepository, error) {
	if store == nil || store.Queries == nil || store.Pool == nil {
		return nil, errors.New("new photo repository: nil postgres store")
	}

	if s3 == nil || s3.Presign == nil || s3.Client == nil || s3.Bucket == "" {
		return nil, errors.New("new photo repository: invalid s3 client data")
	}

	return &uploadRepository{
		store: store,
		s3:    s3,
	}, nil
}

type presignPutObjectParams struct {
	Key         string
	ContentType string
	ExpiresAt   time.Time
}

func (r *uploadRepository) presignPutObject(ctx context.Context, p presignPutObjectParams) (*v4.PresignedHTTPRequest, error) {
	req, err := r.s3.Presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(r.s3.Bucket),
			Key:         aws.String(p.Key),
			ContentType: aws.String(p.ContentType),
		},
		func(options *s3.PresignOptions) {
			options.Expires = time.Until(p.ExpiresAt)
		})
	if err != nil {
		return nil, fmt.Errorf("presign put object: %w", err)
	}
	return req, nil
}

type createPendingPhotoUploadParams struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	Key         string
	ContentType string
	Size        int64
	PhotoDate   time.Time
	Title       string
	Description string
	ExpiresAt   time.Time
}

func (r *uploadRepository) createPendingPhotoUpload(ctx context.Context, p createPendingPhotoUploadParams) error {
	queryParams := db.CreatePendingPhotoUploadRecordParams{
		ID:                p.ID,
		OwnerUserID:       p.OwnerUserID,
		Bucket:            r.s3.Bucket,
		ObjectKeyOriginal: p.Key,
		ContentType:       p.ContentType,
		Size:              p.Size,
		PhotoDate: pgtype.Date{
			Time:  p.PhotoDate,
			Valid: true,
		},
		ExpiresAt: p.ExpiresAt,
	}

	if strings.TrimSpace(p.Title) != "" {
		queryParams.Title = pgtype.Text{
			String: strings.TrimSpace(p.Title),
			Valid:  true,
		}
	}

	if strings.TrimSpace(p.Description) != "" {
		queryParams.Description = pgtype.Text{
			String: strings.TrimSpace(p.Description),
			Valid:  true,
		}
	}

	err := r.store.Queries.CreatePendingPhotoUploadRecord(ctx, queryParams)
	if err != nil {
		return fmt.Errorf("createPendingPhotoUpload: %w", err)
	}
	return nil
}

type pendingPhotoUploadParams struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
}

type pendingPhotoUploadData struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	Bucket      string
	Key         string
	ContentType string
	Size        int64
	PhotoDate   time.Time
	Status      string
}

func (r *uploadRepository) pendingPhotoUpload(ctx context.Context, p pendingPhotoUploadParams) (*pendingPhotoUploadData, error) {
	row, err := r.store.Queries.MarkPhotoUploadRecordProcessing(ctx, db.MarkPhotoUploadRecordProcessingParams{
		ID:          p.ID,
		OwnerUserID: p.OwnerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("pendingPhotoUpload: %w", err)
	}

	return &pendingPhotoUploadData{
		ID:          row.ID,
		OwnerUserID: row.OwnerUserID,
		Bucket:      row.Bucket,
		Key:         row.ObjectKeyOriginal,
		ContentType: row.ContentType,
		Size:        row.Size,
		PhotoDate:   row.PhotoDate.Time,
		Status:      string(row.Status),
	}, nil
}

type completePendingPhotoUploadParams struct {
	ID                 uuid.UUID
	OwnerUserID        uuid.UUID
	ObjectKeyProcessed []byte
	Width              int32
	Height             int32
}

func (r *uploadRepository) completePendingPhotoUpload(ctx context.Context, p completePendingPhotoUploadParams) error {
	queryParams := db.MarkProcessedPhotoUploadRecordCompletedParams{
		ID:                 p.ID,
		OwnerUserID:        p.OwnerUserID,
		ObjectKeyProcessed: p.ObjectKeyProcessed,
		Width: pgtype.Int4{
			Int32: p.Width,
			Valid: true,
		},
		Height: pgtype.Int4{
			Int32: p.Height,
			Valid: true,
		},
	}

	_, err := r.store.Queries.MarkProcessedPhotoUploadRecordCompleted(ctx, queryParams)
	if err != nil {
		return fmt.Errorf("completePendingPhotoUpload: %w", err)
	}
	return nil
}

type failPendingPhotoUploadParams struct {
	ID          uuid.UUID
	OwnerUserID uuid.UUID
	Reason      string
}

func (r *uploadRepository) failPendingPhotoUpload(ctx context.Context, p failPendingPhotoUploadParams) error {
	_, err := r.store.Queries.MarkProcessedPhotoUploadRecordFailed(ctx, db.MarkProcessedPhotoUploadRecordFailedParams{
		ID:          p.ID,
		OwnerUserID: p.OwnerUserID,
		FailureReason: pgtype.Text{
			String: p.Reason,
			Valid:  true,
		},
	})
	if err != nil {
		return fmt.Errorf("failPendingPhotoUpload: %w", err)
	}
	return nil
}
