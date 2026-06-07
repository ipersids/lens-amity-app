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

func GetUserProfile(s *db.Store, u UserProfileDTO) (*db.GetPublicUserProfileRow, error) {
	user, err := s.Queries.GetPublicUserProfile(context.Background(), u.Username)
	if err != nil {
		return nil, fmt.Errorf("db error: %w %s", ErrorGetUserProfile, err.Error())
	}

	return &user, nil
}
