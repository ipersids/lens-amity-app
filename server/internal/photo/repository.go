package photo

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type photoRepository struct {
	store   *db.Store
	bucket  string
	presign *s3.PresignClient
}

func newPhotoRepository(store *db.Store, bucket string, presign *s3.PresignClient) (*photoRepository, error) {
	if store == nil || store.Queries == nil || store.Pool == nil {
		return nil, errors.New("new photo repository: nil postgres store")
	}

	if presign == nil || bucket == "" {
		return nil, errors.New("new photo repository: invalid s3 client data")
	}

	return &photoRepository{
		store:   store,
		bucket:  bucket,
		presign: presign,
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
	req, err := r.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
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
		Bucket:      r.bucket,
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
