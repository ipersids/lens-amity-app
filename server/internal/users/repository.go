package users

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"lensamity/internal/storage"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type usersRepository struct {
	store *db.Store
	s3    *storage.Client
}

func newUsersRepository(store *db.Store, s3 *storage.Client) (*usersRepository, error) {
	if store == nil || store.Queries == nil || store.Pool == nil {
		return nil, errors.New("new users repository: nil postgres store")
	}

	if s3 == nil || s3.Presign == nil || s3.Client == nil || s3.Bucket == "" {
		return nil, errors.New("new users repository: invalid s3 client data")
	}

	return &usersRepository{
		store: store,
		s3:    s3,
	}, nil
}

type presignURLparams struct {
	Bucket    string
	Key       string
	ExpiresAt time.Time
}

func (r *usersRepository) presignURL(ctx context.Context, p presignURLparams) (*v4.PresignedHTTPRequest, error) {
	req, err := r.s3.Presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:          aws.String(p.Bucket),
		Key:             aws.String(p.Key),
		ResponseExpires: aws.Time(p.ExpiresAt),
	})
	if err != nil {
		return nil, err
	}

	return req, nil
}

type profileParams struct {
	Username string
}

func (r *usersRepository) profile(ctx context.Context, p profileParams) (*db.GetUserProfileRow, error) {
	row, err := r.store.Queries.GetUserProfile(ctx, p.Username)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

func (r *usersRepository) accessProfile(ctx context.Context, p profileParams) (*db.GetUserAccessProfileRow, error) {
	row, err := r.store.Queries.GetUserAccessProfile(ctx, p.Username)
	if err != nil {
		return nil, err
	}

	return &row, nil
}

type photosPageParams struct {
	OwnerID    uuid.UUID
	CursorDate time.Time
	CursorID   uuid.UUID
	Limit      int32
}

func (r *usersRepository) photosPage(ctx context.Context, p photosPageParams) ([]db.Photo, error) {
	var rows []db.Photo
	var err error

	if p.CursorID != uuid.Nil {
		rows, err = r.store.Queries.ListUserPhotosAfterCursor(ctx, db.ListUserPhotosAfterCursorParams{
			UserID: p.OwnerID,
			CursorPhotoDate: pgtype.Date{
				Time:  p.CursorDate,
				Valid: true,
			},
			CursorID:   p.CursorID,
			LimitCount: p.Limit,
		})
	} else {
		rows, err = r.store.Queries.ListUserPhotosFirstPage(ctx, db.ListUserPhotosFirstPageParams{
			UserID:     p.OwnerID,
			LimitCount: p.Limit,
		})
	}

	if err != nil {
		return nil, err
	}

	return rows, nil
}

type updateProfileParams struct {
	OwnerID     uuid.UUID
	DisplayName string
	About       string
	Visibility  string
}

func (r *usersRepository) updateProfile(ctx context.Context, p updateProfileParams) error {
	_, err := r.store.Queries.UpdateUserProfile(ctx, db.UpdateUserProfileParams{
		DisplayName: p.DisplayName,
		About:       pgtype.Text{String: p.About, Valid: p.About != ""},
		Visibility:  p.Visibility,
		ID:          p.OwnerID,
	})
	if err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	return nil
}

func (r *usersRepository) updateUsername(ctx context.Context, userID uuid.UUID, newUsername string) (*db.UpdateUsernameRow, error) {
	row, err := r.store.Queries.UpdateUsername(ctx, db.UpdateUsernameParams{ID: userID, NewUsernameKey: newUsername})
	if err != nil {
		return nil, fmt.Errorf("update username: %w", err)
	}

	return &row, nil
}

func (r *usersRepository) deleteProfile(ctx context.Context, userID uuid.UUID) error {
	err := r.store.Queries.DeleteProfile(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *usersRepository) photosKeys(ctx context.Context, userID uuid.UUID) (map[string][]types.ObjectIdentifier, error) {
	res := make(map[string][]types.ObjectIdentifier)
	rows, err := r.store.Queries.ListUserAllImages(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, nil
		}
		return nil, err
	}

	for _, row := range rows {
		res[row.Bucket] = append(res[row.Bucket], types.ObjectIdentifier{Key: &row.ObjectKeyOriginal})
	}

	return res, nil
}

func (r *usersRepository) deleteObjects(
	ctx context.Context,
	bucket string,
	objects []types.ObjectIdentifier,
) error {
	if len(objects) == 0 {
		return nil
	}

	if len(objects) > 1000 {
		return fmt.Errorf("S3 DeleteObjects accepts at most 1000 objects, got %d", len(objects))
	}

	delOut, err := r.s3.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("delete objects from bucket %q: %w", bucket, err)
	}

	if len(delOut.Errors) == 0 {
		return nil
	}

	deleteErrors := make([]error, 0, len(delOut.Errors))

	for _, deleteErr := range delOut.Errors {
		key := aws.ToString(deleteErr.Key)
		code := aws.ToString(deleteErr.Code)
		message := aws.ToString(deleteErr.Message)

		slog.Error(
			"S3 could not delete object",
			"bucket", bucket,
			"key", key,
			"code", code,
			"message", message,
		)

		deleteErrors = append(deleteErrors, fmt.Errorf("%s: %s", key, message))
	}

	return fmt.Errorf(
		"could not delete %d object(s) from bucket %q: %w",
		len(deleteErrors),
		bucket,
		errors.Join(deleteErrors...),
	)
}
