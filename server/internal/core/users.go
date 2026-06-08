package core

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
)

type UserService struct {
	store *db.Store
}

func NewUserService(s *db.Store) *UserService {
	return &UserService{
		store: s,
	}
}

var (
	ErrorGetUserProfile = errors.New("")
)

func (s *UserService) GetUserProfile(ctx context.Context, username string) (*db.GetPublicUserProfileRow, error) {
	user, err := s.store.Queries.GetPublicUserProfile(ctx, username)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("db error: %w %s", ErrorGetUserProfile, err.Error())
	}

	return &user, nil
}
