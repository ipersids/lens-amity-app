package core

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
)

type UserProfileDTO struct {
	Username string
}

var (
	ErrorGetUserProfile = errors.New("")
)

func GetUserProfile(ctx context.Context, s *db.Store, u UserProfileDTO) (*db.GetPublicUserProfileRow, error) {
	user, err := s.Queries.GetPublicUserProfile(ctx, u.Username)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, fmt.Errorf("db error: %w %s", ErrorGetUserProfile, err.Error())
	}

	return &user, nil
}
