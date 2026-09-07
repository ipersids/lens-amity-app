package users

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"lensamity/internal/storage"
	"log/slog"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserService struct {
	repo usersStore
}

type usersStore interface {
	avatarURL(ctx context.Context, p avatarURLparams) (*v4.PresignedHTTPRequest, error)
	profile(ctx context.Context, p profileParams) (*db.GetUserProfileRow, error)
}

func NewUserService(store *db.Store, s3Client *storage.Client) (*UserService, error) {
	if store == nil || s3Client == nil {
		return nil, errors.New("new users service: nil postgres or s3 data")
	}

	repo, err := newUsersRepository(store, s3Client)
	if err != nil {
		return nil, err
	}

	return &UserService{repo: repo}, nil
}

var (
	ErrorGetUserProfile = errors.New("user profile not found")
)

type GetUserProfileResult struct {
	ID                     uuid.UUID
	Username               string
	DisplayName            string
	About                  string
	PhotoCount             int64
	Visibility             string
	JoinedAt               time.Time
	AvatarPresignedRequest *v4.PresignedHTTPRequest
}

func (s *UserService) GetUserProfile(ctx context.Context, username string) (*GetUserProfileResult, error) {
	profile, err := s.repo.profile(ctx, profileParams{Username: username})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %w", ErrorGetUserProfile, err)
		}
		return nil, fmt.Errorf("get user profile: %w", err)
	}

	result := &GetUserProfileResult{
		ID:          profile.ID,
		Username:    profile.UsernameKey,
		DisplayName: profile.UsernameDisplay,
		PhotoCount:  profile.PhotoCount,
		Visibility:  profile.ProfileVisibility,
		JoinedAt:    profile.JoinedAt.Time,
	}

	if profile.About.Valid {
		result.About = profile.About.String
	}

	if !profile.AvatarObjectKey.Valid || !profile.AvatarBucket.Valid {
		return result, nil
	}

	avatarReq, err := s.repo.avatarURL(ctx, avatarURLparams{
		Bucket: profile.AvatarBucket.String,
		Key:    profile.AvatarObjectKey.String,
	})
	if err != nil {
		if storage.IsObjectNotFound(err) {
			slog.Error("user profile avatar object not found", "error", err)
			return result, nil
		}
		return nil, fmt.Errorf("get user profile avatar url: %w", err)
	}

	result.AvatarPresignedRequest = avatarReq

	return result, nil
}
