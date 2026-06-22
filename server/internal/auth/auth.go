package auth

import (
	"context"
	"errors"
	"fmt"
	"lensamity/internal/db"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/text/unicode/norm"
)

// Docs:
// - NIST, SP 800-63B, authentication assurance: https://pages.nist.gov/800-63-4/sp800-63b.html
// - Unicode, Technical Standard #39, https://www.unicode.org/reports/tr39/#Restriction_Level_Detection

// Idle timeout: compare now() with last_seen_at.
// Absolute timeout: compare now() with absolute_expires_at.
type Config struct {
	SessionSecret   string
	IdleTimeout     time.Duration
	AbsoluteTimeout time.Duration
	TouchInterval   time.Duration
}

type AuthService struct {
	config *Config
	store  *db.Store
	tokens *sessionTokens
}

func NewAuthService(store *db.Store, confAuth *Config) *AuthService {
	return &AuthService{
		config: confAuth,
		store:  store,
		tokens: newSessionTokens(confAuth.SessionSecret),
	}
}

var (
	ErrUsernameTaken      = errors.New("username is not available")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSession     = errors.New("invalid session")
	ErrInternal           = errors.New("internal auth error")
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
	ukey := normKey(username)
	udisplay := normDisplay(displayName)

	if err := validateUsernameKey(ukey); err != nil {
		return nil, err
	}

	if err := validatePassword(p, []string{ukey, udisplay}); err != nil {
		return nil, err
	}

	if udisplay == "" {
		udisplay = normDisplay(username)
	}

	if err := validateNameLength(udisplay); err != nil {
		return nil, err
	}

	hash, err := generateFromPassword([]byte(p))
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
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("%w: create user: %w", ErrInternal, err)
	}

	return &SignupResponse{
		UsernameKey:     user.UsernameKey,
		UsernameDisplay: user.UsernameDisplay,
	}, nil
}

type LoginResult struct {
	Username        string
	DisplayName     string
	CookieToken     string
	CookieExpiredAt time.Time
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
			return nil, fmt.Errorf("%w: get full user data by key timeout: %w", ErrInternal, err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			// Keep timing close to a wrong password.
			_ = compareHashAndPassword(dummyHash, []byte(p))
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: get full user data by key: %w", ErrInternal, err)
	}

	if err := compareHashAndPassword(uPrivate.PasswordHash, []byte(p)); err != nil {
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

type SessionResult struct {
	UserID uuid.UUID
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

	return &SessionResult{UserID: session.UserID}, nil
}

func sessionIsActive(session db.Session, now time.Time, idleTimeout time.Duration) bool {
	return !session.RevokedAt.Valid &&
		now.Before(session.AbsoluteExpiresAt) &&
		now.Before(session.LastSeenAt.Add(idleTimeout))
}
