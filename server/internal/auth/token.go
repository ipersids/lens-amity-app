package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	tokenIssuer       = "lensamity-app"
	tokenAudienceUser = "USER"
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token has expired")
)

func (s *AuthService) signAccessToken(ctx context.Context, userID uuid.UUID, nowUTC time.Time) (string, error) {
	accessTokenClaims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   userID.String(),
		Audience:  jwt.ClaimStrings{tokenAudienceUser},
		ExpiresAt: jwt.NewNumericDate(nowUTC.Add(s.conf.JWTexpiry)),
		IssuedAt:  jwt.NewNumericDate(nowUTC),
		NotBefore: jwt.NewNumericDate(nowUTC),
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims).SignedString([]byte(s.conf.JWTsecret))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

type signedRefreshTokenData struct {
	id     uuid.UUID
	claims *jwt.RegisteredClaims
	token  string
}

func (s *AuthService) signRefreshToken(ctx context.Context, userID uuid.UUID, nowUTC time.Time) (*signedRefreshTokenData, error) {
	refreshTokenUUID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	refreshTokenClaims := jwt.RegisteredClaims{
		ID:        refreshTokenUUID.String(),
		Issuer:    tokenIssuer,
		Subject:   userID.String(),
		Audience:  jwt.ClaimStrings{tokenAudienceUser},
		ExpiresAt: jwt.NewNumericDate(nowUTC.Add(s.conf.RefreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(nowUTC),
		NotBefore: jwt.NewNumericDate(nowUTC),
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims).SignedString([]byte(s.conf.RefreshSecret))
	if err != nil {
		return nil, err
	}

	return &signedRefreshTokenData{
		id:     refreshTokenUUID,
		claims: &refreshTokenClaims,
		token:  refreshToken,
	}, nil
}

func validateToken(tokenStr string, base string) (*jwt.RegisteredClaims, error) {
	var claims jwt.RegisteredClaims

	parsedToken, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (any, error) {
		return []byte(base), nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tokenIssuer),
		jwt.WithAudience(tokenAudienceUser),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if parsedToken.Valid {
		return &claims, nil
	}

	return nil, ErrInvalidToken
}
