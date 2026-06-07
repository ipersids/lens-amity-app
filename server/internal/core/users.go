package core

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserProfileDTO struct {
	UUID string
}

var (
	ErrorGetUserProfile = errors.New("")
)

func GetUserProfile(s *db.Store, u UserProfileDTO) (*db.GetUserRow, error) {
	userUUID, err := uuid.Parse(u.UUID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid: %w %s", ErrorGetUserProfile, err.Error())
	}

	user, err := s.Queries.GetUser(context.Background(), pgtype.UUID{Bytes: userUUID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("db error: %w %s", ErrorGetUserProfile, err.Error())
	}

	return &user, nil
}
