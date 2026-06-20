package auth

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	ErrUsernameTaken      = errors.New("username is not available")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrCompromisedToken   = errors.New("token was compromised")
	ErrInternal           = errors.New("internal auth error")
)

const dummyHash = "$argon2id$v=19$m=65536,t=3,p=2$72aaaaK2bbDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA"

type SignupResponse struct {
	UsernameKey     string
	UsernameDisplay string
}

func (s *AuthService) Signup(ctx context.Context, uername, displayName, password string) (*SignupResponse, error) {
	p := norm.NFC.String(password)
	ukey := normKey(uername)
	udisplay := normDisplay(displayName)

	if err := validateUsernameKey(ukey); err != nil {
		return nil, err
	}

	if err := validatePassword(p, []string{ukey, udisplay}); err != nil {
		return nil, err
	}

	if udisplay == "" {
		udisplay = normDisplay(uername)
	}

	if err := validateNameLength(udisplay); err != nil {
		return nil, err
	}

	hash, err := generateFromPassword([]byte(p))
	if err != nil {
		return nil, fmt.Errorf("%w: generate password hash: %v", ErrInternal, err)
	}

	user, err := s.store.Queries.CreateUser(ctx, db.CreateUserParams{
		UsernameKey:     ukey,
		PasswordHash:    hash,
		UsernameDisplay: udisplay,
	})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create user timeout: %w", err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // PostgreSQL Error Codes: "23505" - unique_violation
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("%w: create user: %v", ErrInternal, err)
	}

	return &SignupResponse{
		UsernameKey:     user.UsernameKey,
		UsernameDisplay: user.UsernameDisplay,
	}, nil
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
		return nil, ErrInvalidCredentials
	}

	uPrivate, err := s.store.Queries.GetFullUserDataByKey(ctx, ukey)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get full user data by key timeout: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			// Keep timing close to a wrong password.
			_ = compareHashAndPassword(dummyHash, []byte(p))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: get full user data by key: %v", ErrInternal, err)
	}

	if err := compareHashAndPassword(uPrivate.PasswordHash, []byte(p)); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: compare hash and password: %v", ErrInternal, err)
	}

	nowUTC := time.Now().UTC()

	token, err := s.conf.signAccessToken(ctx, uPrivate.ID, nowUTC)
	if err != nil {
		return nil, fmt.Errorf("%w: sign access token: %v", ErrInternal, err)
	}
	refreshTokenData, err := s.conf.signRefreshToken(ctx, uPrivate.ID, nowUTC)
	if err != nil {
		return nil, fmt.Errorf("%w: sign refresh token: %v", ErrInternal, err)
	}

	err = s.store.Queries.CreateNewRefreshToken(ctx, db.CreateNewRefreshTokenParams{
		ID:        refreshTokenData.id,
		UserID:    uPrivate.ID,
		ExpiresAt: refreshTokenData.claims.ExpiresAt.Time,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create new refresh token timeout: %w", err)
		}
		return nil, fmt.Errorf("%w: create new refresh token: %v", ErrInternal, err)
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
		return nil, ErrInvalidToken
	}

	now := time.Now().Truncate(time.Second)

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("begin refresh transaction timeout: %w", err)
		}
		return nil, fmt.Errorf("%w: begin refresh transaction: %v", ErrInternal, err)
	}
	defer tx.Rollback(ctx)

	rtoken, err := s.store.Queries.WithTx(tx).GetRefreshTokenForUpdate(ctx, db.GetRefreshTokenForUpdateParams{ID: tokenID, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get refresh token for update timeout: %w", err)
		}
		return nil, fmt.Errorf("%w: get refresh token for update: %v", ErrInternal, err)
	}

	if rtoken.Revoked {
		if rtoken.GracePeriodUntil.Valid && time.Now().Before(rtoken.GracePeriodUntil.Time) {
			return &RefreshResult{
				Replayed:     true,
				AccessToken:  "",
				RefreshToken: "",
			}, nil
		}
		// Enable reuse detection
		if _, err := s.store.Queries.WithTx(tx).RevokeAllUserTokens(ctx, db.RevokeAllUserTokensParams{
			UserID:        userID,
			RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonReplayed, Valid: true},
		}); err != nil {
			return nil, fmt.Errorf("%w: revoke replayed user tokens: %v", ErrInternal, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("%w: commit replayed token revocation: %v", ErrInternal, err)
		}
		return nil, ErrCompromisedToken
	}

	accessToken, err := s.conf.signAccessToken(ctx, userID, now)
	if err != nil {
		return nil, fmt.Errorf("%w: sign access token: %v", ErrInternal, err)
	}

	refreshTokenData, err := s.conf.signRefreshToken(ctx, userID, now)
	if err != nil {
		return nil, fmt.Errorf("%w: sign refresh token: %v", ErrInternal, err)
	}

	_, err = s.store.Queries.WithTx(tx).RotateRefreshToken(ctx, db.RotateRefreshTokenParams{
		ID:               tokenID,
		UserID:           userID,
		GracePeriodUntil: pgtype.Timestamptz{Time: now.Add(3 * time.Second), Valid: true},
		RevokedReason:    db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonRefresh, Valid: true},
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("rotate refresh token timeout: %w", err)
		}
		return nil, fmt.Errorf("%w: rotate refresh token: %v", ErrInternal, err)
	}

	err = s.store.Queries.WithTx(tx).CreateNewRefreshToken(ctx, db.CreateNewRefreshTokenParams{
		ID:        refreshTokenData.id,
		UserID:    userID,
		ExpiresAt: refreshTokenData.claims.ExpiresAt.Time,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create rotated refresh token timeout: %w", err)
		}
		return nil, fmt.Errorf("%w: create rotated refresh token: %v", ErrInternal, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: commit refresh token rotation: %v", ErrInternal, err)
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
		return err
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return ErrInvalidToken
	}

	tokenID, err := uuid.Parse(claims.ID)
	if err != nil {
		return ErrInvalidToken
	}

	_, err = s.store.Queries.RevokeRefreshToken(ctx, db.RevokeRefreshTokenParams{
		RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonLogout, Valid: true},
		ID:            tokenID,
		UserID:        userID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("revoke refresh token timeout: %w", err)
		}
		return fmt.Errorf("%w: revoke refresh token: %v", ErrInternal, err)
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	_, err := s.store.Queries.RevokeAllUserTokens(ctx, db.RevokeAllUserTokensParams{
		UserID:        userID,
		RevokedReason: db.NullTokenRevokedReason{TokenRevokedReason: db.TokenRevokedReasonLogout, Valid: true},
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("revoke all user tokens timeout: %w", err)
		}
		return fmt.Errorf("%w: revoke all user tokens: %v", ErrInternal, err)
	}
	return nil
}
