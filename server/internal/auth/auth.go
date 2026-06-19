package auth

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/unicode/norm"
)

// Docs:
// - NIST, SP 800-63B, authentication assurance: https://pages.nist.gov/800-63-4/sp800-63b.html
// - Unicode, Technical Standard #39, https://www.unicode.org/reports/tr39/#Restriction_Level_Detection

type Config struct {
	JWTsecret     string
	RefreshSecret string
	JWTexpiry     time.Duration
	RefreshExpiry time.Duration
}

type AuthService struct {
	conf  *Config
	store *db.Store
}

func NewAuthService(s *db.Store, confAuth *Config) *AuthService {
	return &AuthService{
		conf:  confAuth,
		store: s,
	}
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	return validateToken(tokenStr, s.conf.JWTsecret)
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCompromisedToken   = errors.New("token was compromised")
)

const dummyHash = "$argon2id$v=19$m=65536,t=3,p=2$72aaaaK2bbDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA"

func (s *AuthService) Signup(ctx context.Context, uername, displayName, password string) (*db.CreateUserRow, error) {
	p := norm.NFC.String(password)
	ukey := normKey(uername)
	udisplay := normDisplay(displayName)

	if err := validateUsernameKey(ukey); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCredentials, err.Error())
	}

	if err := validatePassword(p, []string{ukey, udisplay}); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCredentials, err.Error())
	}

	if udisplay == "" {
		udisplay = normDisplay(uername)
	}

	if err := validateNameLength(udisplay); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidCredentials, err.Error())
	}

	hash, err := generateFromPassword([]byte(p))
	if err != nil {
		return nil, err
	}

	user, err := s.store.Queries.CreateUser(ctx, db.CreateUserParams{
		UsernameKey:     ukey,
		PasswordHash:    hash,
		UsernameDisplay: udisplay,
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, ErrInvalidCredentials
	}

	return &user, nil
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	Username     string
	DisplayName  string
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	p := norm.NFC.String(password)
	ukey := normKey(username)

	if err := validatePasswordLength(p); err != nil {
		return nil, err
	}

	uPrivate, err := s.store.Queries.GetFullUserDataByKey(ctx, ukey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = compareHashAndPassword(dummyHash, []byte(p))
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := compareHashAndPassword(uPrivate.PasswordHash, []byte(p)); err != nil {
		return nil, err
	}

	nowUTC := time.Now().UTC()

	token, err := s.conf.signAccessToken(ctx, uPrivate.ID, nowUTC)
	if err != nil {
		return nil, err
	}
	refreshTokenData, err := s.conf.signRefreshToken(ctx, uPrivate.ID, nowUTC)
	if err != nil {
		return nil, err
	}

	err = s.store.Queries.CreateNewRefreshToken(ctx, db.CreateNewRefreshTokenParams{
		ID:        refreshTokenData.id,
		UserID:    uPrivate.ID,
		ExpiresAt: refreshTokenData.claims.ExpiresAt.Time,
	})
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  token,
		RefreshToken: refreshTokenData.token,
		Username:     uPrivate.UsernameKey,
		DisplayName:  uPrivate.UsernameDisplay,
	}, nil
}

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
	Replayed     bool
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	claims, err := validateToken(refreshToken, s.conf.RefreshSecret)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, err
	}

	now := time.Now().Truncate(time.Second)

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, err
	}

	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	rtoken, err := s.store.Queries.WithTx(tx).GetRefreshTokenForUpdate(ctx, db.GetRefreshTokenForUpdateParams{ID: tokenID, UserID: userID})
	if err != nil {
		return nil, err
	}

	if rtoken.Revoked {
		if rtoken.GracePeriodUntil.Valid && time.Now().Before(rtoken.GracePeriodUntil.Time) {
			return &RefreshResult{
				Replayed:     true,
				AccessToken:  "",
				RefreshToken: "",
			}, nil
		}
		// enable reuse detection
		if _, err := s.store.Queries.WithTx(tx).RevokeAllUserTokens(ctx, db.RevokeAllUserTokensParams{
			UserID:        userID,
			RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonReplayed, Valid: true},
		}); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCompromisedToken, err.Error())
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("%w: %s", ErrCompromisedToken, err.Error())
		}
		return nil, ErrCompromisedToken
	}

	accessToken, err := s.conf.signAccessToken(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	refreshTokenData, err := s.conf.signRefreshToken(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	_, err = s.store.Queries.WithTx(tx).RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		ID:               tokenID,
		UserID:           userID,
		GracePeriodUntil: pgtype.Timestamptz{Time: now.Add(3 * time.Second), Valid: true},
		RevokedReason:    db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonRefresh, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	err = s.store.Queries.WithTx(tx).CreateNewRefreshToken(ctx, db.CreateNewRefreshTokenParams{
		ID:        refreshTokenData.id,
		UserID:    userID,
		ExpiresAt: refreshTokenData.claims.ExpiresAt.Time,
	})
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &RefreshResult{
		Replayed:     false,
		AccessToken:  accessToken,
		RefreshToken: refreshTokenData.token,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := validateToken(refreshToken, s.conf.RefreshSecret)
	if err != nil {
		slog.Error("1", "", err)
		return nil
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		slog.Error("2", "", err)
		return err
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return err
	}

	_, err = s.store.Queries.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{
		RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonLogout, Valid: true},
		ID:            tokenID,
		UserID:        userID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	_, err := s.store.Queries.RevokeAllUserTokens(ctx, db.RevokeAllUserTokensParams{
		UserID:        userID,
		RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonLogout, Valid: true},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return nil
}
