package photo

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"lensamity/internal/storage"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type photoRepository struct {
	store *db.Store
	s3    *storage.Client
}

func newPhotoRepository(store *db.Store, s3 *storage.Client) (*photoRepository, error) {
	if store == nil || store.Queries == nil || store.Pool == nil {
		return nil, errors.New("new photo repository: nil postgres store")
	}

	if s3 == nil || s3.Presign == nil || s3.Client == nil || s3.Bucket == "" {
		return nil, errors.New("new photo repository: invalid s3 client data")
	}

	return &photoRepository{
		store: store,
		s3:    s3,
	}, nil
}

type createPendingUploadParams struct {
	PhotoID     uuid.UUID
	OwnerUserID uuid.UUID
	ObjectKey   string
	LocalDate   time.Time
	ContentType string
	ExpiresAt   time.Time
}

func (r *photoRepository) createPendingUpload(ctx context.Context, p createPendingUploadParams) (*UploadObjectRequest, error) {
	req, err := r.s3.Presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.s3.Bucket),
		Key:         aws.String(p.ObjectKey),
		ContentType: aws.String(p.ContentType),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Until(p.ExpiresAt)
	})
	if err != nil {
		return nil, fmt.Errorf("presign put object: %w", err)
	}

	err = r.store.Queries.CreatePendingPhotoRecord(ctx, db.CreatePendingPhotoRecordParams{
		ID:          p.PhotoID,
		OwnerUserID: p.OwnerUserID,
		Bucket:      r.s3.Bucket,
		ObjectKey:   p.ObjectKey,
		LocalDate: pgtype.Date{
			Time:  p.LocalDate,
			Valid: true,
		},
		UploadExpiresAt: p.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create pending photo record: %w", err)
	}

	return &UploadObjectRequest{
		PhotoID: p.PhotoID,
		URL:     req.URL,
		Method:  req.Method,
		Header:  req.SignedHeader,
	}, nil
}

type UploadObjectStatus struct {
	Status        db.PhotoProcessingStatus
	ContentType   *string
	ContentLength *int64
}

func (r *photoRepository) getUploadStatus(ctx context.Context, photoID uuid.UUID, userID uuid.UUID) (*UploadObjectStatus, error) {
	record, err := r.store.Queries.GetPhotoRecord(ctx, db.GetPhotoRecordParams{ID: photoID, OwnerUserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPhotoNotFound
		}
		return nil, fmt.Errorf("get photo record: %w", err)
	}
	if record.Status != db.PhotoProcessingStatusPending {
		return &UploadObjectStatus{Status: record.Status}, nil
	}

	head, err := r.s3.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(record.Bucket),
		Key:    aws.String(record.ObjectKey),
	})
	if err != nil {
		var notFound *s3types.NotFound
		var noSuchKey *s3types.NoSuchKey
		if errors.As(err, &notFound) || errors.As(err, &noSuchKey) {
			return nil, ErrUploadNotFound
		}
		return nil, fmt.Errorf("get head object record: %w", err)
	}

	return &UploadObjectStatus{
		Status:        record.Status,
		ContentType:   head.ContentType,
		ContentLength: head.ContentLength,
	}, nil
}

func (r *photoRepository) setCompleteUploadStatus(ctx context.Context, photoID uuid.UUID, userID uuid.UUID) (bool, error) {
	_, err := r.store.Queries.MarkPhotoReady(ctx, db.MarkPhotoReadyParams{
		ID:          photoID,
		OwnerUserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("set photo status ready: %w", err)
	}
	return true, nil
}
