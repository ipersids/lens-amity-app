package auth

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/unicode/norm"
)

const (
	idleTimeout     = 36 * time.Hour
	AbsoluteTimeout = 7 * 24 * time.Hour
	touchInterval   = 15 * time.Minute
)

// Docs:
// - NIST, SP 800-63B, authentication assurance: https://pages.nist.gov/800-63-4/sp800-63b.html
// - Unicode, Technical Standard #39, https://www.unicode.org/reports/tr39/#Restriction_Level_Detection

// Idle timeout: compare now() with last_seen_at.
// Absolute timeout: compare now() with absolute_expires_at.
type Config struct {
	IdleTimeout     time.Duration
	AbsoluteTimeout time.Duration
	TouchInterval   time.Duration
}

type AuthService struct {
	config Config
	store  *db.Store
	tokens sessionTokens
}

func NewAuthService(store *db.Store, sessionSecret string) (*AuthService, error) {
	if store == nil {
		return nil, errors.New("auth: nil store")
	}
	if strings.TrimSpace(sessionSecret) == "" {
		return nil, errors.New("auth: session secret is required")
	}

	return &AuthService{
		config: Config{
			IdleTimeout:     idleTimeout,
			AbsoluteTimeout: AbsoluteTimeout,
			TouchInterval:   touchInterval,
		},
		store:  store,
		tokens: newSessionTokens(sessionSecret),
	}, nil
}

var (
	ErrUsernameUnavailable      = errors.New("username is not available")
	ErrUsernameValidationFailed = errors.New("invalid username")
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrNewPasswordValidation    = errors.New("invalid new password")
	ErrPasswordChanged          = errors.New("password changed in another session")
	ErrInvalidSession           = errors.New("invalid session")
	ErrInternal                 = errors.New("internal error")
)

const (
	dummyHash                   = "$argon2id$v=19$m=65536,t=3,p=2$72aaaaK2bbDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA"
	usernameKeyUniqueConstraint = "users_username_key_key"
)

type SignupResponse struct {
	UsernameKey     string
	UsernameDisplay string
}

func (s *AuthService) Signup(ctx context.Context, username, displayName, password string) (*SignupResponse, error) {
	p := norm.NFC.String(password)
	ukey := NormKey(username)
	udisplay := NormText(displayName)

	if err := ValidateUsernameKey(ukey); err != nil {
		return nil, err
	}

	if err := ValidatePassword(p, []string{ukey}); err != nil {
		return nil, err
	}

	if udisplay == "" {
		udisplay = NormText(username)
	}

	if err := ValidateNameLength(udisplay); err != nil {
		return nil, err
	}

	hash, err := GenerateFromPassword([]byte(p))
	if err != nil {
		return nil, fmt.Errorf("%w: generate password hash: %w", ErrInternal, err)
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
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == usernameKeyUniqueConstraint {
			return nil, ErrUsernameUnavailable
		}
		return nil, fmt.Errorf("%w: create user: %w", ErrInternal, err)
	}

	return &SignupResponse{
		UsernameKey:     user.UsernameKey,
		UsernameDisplay: user.UsernameDisplay,
	}, nil
}

func (s *AuthService) UsernameExists(ctx context.Context, username string) (bool, error) {
	ukey := NormKey(username)

	if err := ValidateUsernameKey(ukey); err != nil {
		if errors.Is(err, ErrUsernameUnavailable) {
			return false, nil
		}
		return false, fmt.Errorf("%w: %w", ErrUsernameValidationFailed, err)
	}

	exists, err := s.store.Queries.UsernameExists(ctx, ukey)
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrInternal, err)
	}

	return !exists, nil
}

type LoginResult struct {
	Username        string
	DisplayName     string
	CookieToken     string
	CookieExpiredAt time.Time
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	p := norm.NFC.String(password)
	ukey := NormKey(username)

	if err := validatePasswordLength(p); err != nil {
		return nil, ErrInvalidCredentials
	}

	uPrivate, err := s.store.Queries.GetUserDataForLogin(ctx, ukey)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: get full user data by key timeout: %w", ErrInternal, err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			// Keep timing close to a wrong password.
			_ = CompareHashAndPassword(dummyHash, []byte(p))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: get full user data by key: %w", ErrInternal, err)
	}

	if err := CompareHashAndPassword(uPrivate.PasswordHash, []byte(p)); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: compare hash and password: %w", ErrInternal, err)
	}

	cookieCredentials, err := s.tokens.New()
	if err != nil {
		return nil, fmt.Errorf("%w: create cookie credentials: %w", ErrInternal, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.config.AbsoluteTimeout)
	_, err = s.store.Queries.CreateSession(ctx, db.CreateSessionParams{
		TokenHash:         cookieCredentials.hash,
		UserID:            uPrivate.ID,
		CreatedAt:         now,
		LastSeenAt:        now,
		AbsoluteExpiresAt: expiresAt,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create session timeout: %w", err)
		}
		return nil, fmt.Errorf("%w: create session: %w", ErrInternal, err)
	}

	return &LoginResult{
		Username:        uPrivate.UsernameKey,
		DisplayName:     uPrivate.UsernameDisplay,
		CookieToken:     cookieCredentials.cookie,
		CookieExpiredAt: expiresAt,
	}, nil
}

type UpdatePasswordParams struct {
	UserID              uuid.UUID
	Username            string
	CurrentPassword     string
	NewPassword         string
	CurrentSessionToken string
	RevokeAll           bool
}

type UpdatePasswordResult struct {
	CookieToken     string
	CookieExpiredAt time.Time
}

func (s *AuthService) UpdatePassword(ctx context.Context, p UpdatePasswordParams) (*UpdatePasswordResult, error) {
	currentPassword := norm.NFC.String(p.CurrentPassword)
	newPassword := norm.NFC.String(p.NewPassword)

	if currentPassword == newPassword {
		return nil, fmt.Errorf("%w: old and new passwords should be different", ErrNewPasswordValidation)
	}

	user, err := s.store.Queries.GetPasswordHash(ctx, p.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = CompareHashAndPassword(dummyHash, []byte(currentPassword))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: get password hash: %w", ErrInternal, err)
	}

	if err := CompareHashAndPassword(user.PasswordHash, []byte(currentPassword)); err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: compare password hash: %w", ErrInternal, err)
	}

	if err := ValidatePassword(newPassword, []string{NormKey(p.Username)}); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrNewPasswordValidation, err)
	}

	newPasswordHash, err := GenerateFromPassword([]byte(newPassword))
	if err != nil {
		return nil, fmt.Errorf("%w: generate password hash: %w", ErrInternal, err)
	}

	currentTokenHash, err := s.tokens.Hash(p.CurrentSessionToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidSession, err)
	}

	newCredentials, err := s.tokens.New()
	if err != nil {
		return nil, fmt.Errorf("%w: create session credentials: %w", ErrInternal, err)
	}

	tx, err := s.store.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: begin password update transaction: %w", ErrInternal, err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	queries := s.store.Queries.WithTx(tx)
	_, err = queries.UpdatePasswordHash(ctx, db.UpdatePasswordHashParams{
		UserID:              p.UserID,
		CurrentPasswordHash: user.PasswordHash,
		NewPasswordHash:     newPasswordHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPasswordChanged
		}
		return nil, fmt.Errorf("%w: update password hash: %w", ErrInternal, err)
	}

	revokedAt := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	switch {
	case p.RevokeAll:
		err = queries.RevokeAllSessions(ctx, db.RevokeAllSessionsParams{
			RevokedAt: revokedAt,
			UserID:    p.UserID,
		})
	default:
		_, err = queries.RevokeSession(ctx, db.RevokeSessionParams{
			RevokedAt: revokedAt,
			TokenHash: currentTokenHash,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidSession
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%w: revoke sessions: %w", ErrInternal, err)
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.config.AbsoluteTimeout)
	_, err = queries.CreateSession(ctx, db.CreateSessionParams{
		TokenHash:         newCredentials.hash,
		UserID:            p.UserID,
		CreatedAt:         now,
		LastSeenAt:        now,
		AbsoluteExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: create replacement session: %w", ErrInternal, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%w: commit password update transaction: %w", ErrInternal, err)
	}

	return &UpdatePasswordResult{
		CookieToken:     newCredentials.cookie,
		CookieExpiredAt: expiresAt,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, cookie string) error {
	tokenHash, err := s.tokens.Hash(cookie)
	if err != nil {
		if errors.Is(err, ErrInvalidSessionToken) {
			return nil
		}
		return fmt.Errorf("%w: hash session token: %w", ErrInternal, err)
	}

	_, err = s.store.Queries.RevokeSession(ctx, db.RevokeSessionParams{
		RevokedAt: pgtype.Timestamptz{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		TokenHash: tokenHash,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: revoke session: %w", ErrInternal, err)
		}
		return fmt.Errorf("%w: revoke session: %w", ErrInternal, err)
	}

	return nil
}

func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	err := s.store.Queries.RevokeAllSessions(ctx, db.RevokeAllSessionsParams{
		RevokedAt: pgtype.Timestamptz{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: revoke all sessions: %w", ErrInternal, err)
		}
		return fmt.Errorf("%w: revoke all sessions: %w", ErrInternal, err)
	}

	return nil
}

type SessionOwnerResult struct {
	UsernameKey     string
	UsernameDisplay string
}

func (s *AuthService) SessionOwner(ctx context.Context, userID uuid.UUID) (*SessionOwnerResult, error) {
	user, err := s.store.Queries.GetSessionOwner(ctx, userID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: get session owner data by id timeout: %w", ErrInternal, err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: session owner data by id: %w", ErrInternal, err)
	}

	return &SessionOwnerResult{UsernameKey: user.UsernameKey, UsernameDisplay: user.UsernameDisplay}, nil
}

type SessionResult struct {
	Username string
	UserID   uuid.UUID
}

func (s *AuthService) ValidateSession(ctx context.Context, cookie string) (*SessionResult, error) {
	tokenHash, err := s.tokens.Hash(cookie)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidSession, err)
	}

	session, err := s.store.Queries.GetSession(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidSession
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: validate session timeout: %w", ErrInternal, err)
		}
		return nil, fmt.Errorf("%w: validate session: %w", ErrInternal, err)
	}

	now := time.Now().UTC()
	if !sessionIsActive(session, now, s.config.IdleTimeout) {
		return nil, ErrInvalidSession
	}

	if !now.Before(session.LastSeenAt.Add(s.config.TouchInterval)) {
		_, err = s.store.Queries.UpdateSessionActivity(ctx, db.UpdateSessionActivityParams{
			LastSeenAt: now,
			TokenHash:  tokenHash,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrInvalidSession
			}
			return nil, fmt.Errorf("%w: update session activity: %w", ErrInternal, err)
		}
	}

	return &SessionResult{UserID: session.UserID, Username: session.UsernameKey.String}, nil
}

func sessionIsActive(session db.GetSessionRow, now time.Time, idleTimeout time.Duration) bool {
	return !session.RevokedAt.Valid &&
		now.Before(session.AbsoluteExpiresAt) &&
		now.Before(session.LastSeenAt.Add(idleTimeout))
}
