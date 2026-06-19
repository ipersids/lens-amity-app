package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func testTokenConfig() *Config {
	return &Config{
		JWTsecret:     "test-access-secret",
		RefreshSecret: "test-refresh-secret",
		JWTexpiry:     15 * time.Minute,
		RefreshExpiry: 24 * time.Hour,
	}
}

func TestConfigSignAccessToken(t *testing.T) {
	t.Parallel()

	conf := testTokenConfig()
	userID := uuid.MustParse("018f90a8-0662-7cc7-98f5-5b25774fdd5a")
	now := time.Now().UTC().Truncate(time.Second)

	token, err := conf.signAccessToken(context.Background(), userID, now)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := validateToken(token, conf.JWTsecret)
	if err != nil {
		t.Fatal(err)
	}

	if claims.Subject != userID.String() {
		t.Fatalf("subject = %q, want %q", claims.Subject, userID.String())
	}
	if claims.Issuer != tokenIssuer {
		t.Fatalf("issuer = %q, want %q", claims.Issuer, tokenIssuer)
	}
	if !claimStringsContains(claims.Audience, tokenAudienceUser) {
		t.Fatalf("audience = %v, want %q", claims.Audience, tokenAudienceUser)
	}
	if got, want := claims.ExpiresAt.Time, now.Add(conf.JWTexpiry); !got.Equal(want) {
		t.Fatalf("expires at = %s, want %s", got, want)
	}
}

func TestConfigSignRefreshToken(t *testing.T) {
	t.Parallel()

	conf := testTokenConfig()
	userID := uuid.MustParse("018f90a8-0662-7cc7-98f5-5b25774fdd5a")
	now := time.Now().UTC().Truncate(time.Second)

	tokenData, err := conf.signRefreshToken(context.Background(), userID, now)
	if err != nil {
		t.Fatal(err)
	}

	if tokenData.id == uuid.Nil {
		t.Fatal("refresh token id is nil")
	}
	if tokenData.claims.ID != tokenData.id.String() {
		t.Fatalf("claim id = %q, want %q", tokenData.claims.ID, tokenData.id.String())
	}

	claims, err := validateToken(tokenData.token, conf.RefreshSecret)
	if err != nil {
		t.Fatal(err)
	}

	if claims.ID != tokenData.id.String() {
		t.Fatalf("validated claim id = %q, want %q", claims.ID, tokenData.id.String())
	}
	if claims.Subject != userID.String() {
		t.Fatalf("subject = %q, want %q", claims.Subject, userID.String())
	}
	if got, want := claims.ExpiresAt.Time, now.Add(conf.RefreshExpiry); !got.Equal(want) {
		t.Fatalf("expires at = %s, want %s", got, want)
	}
}

func TestValidateTokenRejectsInvalidTokens(t *testing.T) {
	t.Parallel()

	conf := testTokenConfig()
	userID := uuid.MustParse("018f90a8-0662-7cc7-98f5-5b25774fdd5a")
	now := time.Now().UTC().Truncate(time.Second)

	validClaims := jwt.RegisteredClaims{
		Issuer:    tokenIssuer,
		Subject:   userID.String(),
		Audience:  jwt.ClaimStrings{tokenAudienceUser},
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	tests := []struct {
		name      string
		token     string
		secret    string
		wantError error
	}{
		{
			name:      "malformed token",
			token:     "not-a-jwt",
			secret:    conf.JWTsecret,
			wantError: ErrInvalidToken,
		},
		{
			name:      "wrong secret",
			token:     mustSignTestToken(t, jwt.SigningMethodHS256, validClaims, conf.JWTsecret),
			secret:    "wrong-secret",
			wantError: ErrInvalidToken,
		},
		{
			name: "expired token",
			token: mustSignTestToken(t, jwt.SigningMethodHS256, jwt.RegisteredClaims{
				Issuer:    tokenIssuer,
				Subject:   userID.String(),
				Audience:  jwt.ClaimStrings{tokenAudienceUser},
				ExpiresAt: jwt.NewNumericDate(now.Add(-time.Minute)),
				IssuedAt:  jwt.NewNumericDate(now.Add(-time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(-time.Hour)),
			}, conf.JWTsecret),
			secret:    conf.JWTsecret,
			wantError: ErrExpiredToken,
		},
		{
			name: "wrong issuer",
			token: mustSignTestToken(t, jwt.SigningMethodHS256, jwt.RegisteredClaims{
				Issuer:    "other-app",
				Subject:   userID.String(),
				Audience:  jwt.ClaimStrings{tokenAudienceUser},
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
			}, conf.JWTsecret),
			secret:    conf.JWTsecret,
			wantError: ErrInvalidToken,
		},
		{
			name: "wrong audience",
			token: mustSignTestToken(t, jwt.SigningMethodHS256, jwt.RegisteredClaims{
				Issuer:    tokenIssuer,
				Subject:   userID.String(),
				Audience:  jwt.ClaimStrings{"ADMIN"},
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(now),
				NotBefore: jwt.NewNumericDate(now),
			}, conf.JWTsecret),
			secret:    conf.JWTsecret,
			wantError: ErrInvalidToken,
		},
		{
			name:      "wrong signing method",
			token:     mustSignTestToken(t, jwt.SigningMethodHS512, validClaims, conf.JWTsecret),
			secret:    conf.JWTsecret,
			wantError: ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := validateToken(tt.token, tt.secret)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("validateToken() error = %v, want %v", err, tt.wantError)
			}
		})
	}
}

func mustSignTestToken(t *testing.T, method jwt.SigningMethod, claims jwt.RegisteredClaims, secret string) string {
	t.Helper()

	token, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	return token
}

func claimStringsContains(claims jwt.ClaimStrings, want string) bool {
	for _, claim := range claims {
		if claim == want {
			return true
		}
	}

	return false
}
