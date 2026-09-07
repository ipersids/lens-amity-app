package users

import (
	"context"
	"errors"
	"lensamity/internal/db"
	"lensamity/internal/storage"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

type avatarURLparams struct {
	Bucket string
	Key    string
}

func (r *usersRepository) avatarURL(ctx context.Context, p avatarURLparams) (*v4.PresignedHTTPRequest, error) {
	req, err := r.s3.Presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:          aws.String(p.Bucket),
		Key:             aws.String(p.Key),
		ResponseExpires: aws.Time(time.Now().Add(time.Hour)),
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
